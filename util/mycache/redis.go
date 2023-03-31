package mycache

import (
	"context"
	"fmt"
	"hoyobar/util/funcs"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb *redis.Client
}

var _ Cache = (*RedisCache)(nil)
var _ TimeOrderedSetCache = (*RedisCache)(nil)

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{
		rdb: rdb,
	}
}

// Get implements Cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	res, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	err = errors.Wrapf(err, "fail to get %v from redis", key)
	if err != nil {
		log.Println(err)
	}
	return res, err
}

// Multi get
func (r *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	res, err := r.rdb.MGet(ctx, keys...).Result()
	err = errors.Wrapf(err, "fail to mget %v from redis", keys)
	if err != nil {
		log.Println(err)
	}
	return res, err
}

// Set implements Cache
func (r *RedisCache) Set(ctx context.Context, key string, value string, d time.Duration) error {
	err := r.rdb.Set(ctx, key, value, d).Err()
	err = errors.Wrapf(err, "fail to set value with key %v", key)
	if err != nil {
		log.Println(err)
	}
	return err
}

// GetInt64 implements Cache
func (r *RedisCache) GetInt64(ctx context.Context, key string) (int64, error) {
	cmd := r.rdb.Get(ctx, key)
	err := cmd.Err()
	if err == redis.Nil {
		return 0, ErrNotFound
	}
	if err != nil {
		err = errors.Wrapf(err, "fail to get %v from redis", key)
		log.Println(err)
		return 0, err
	}
	value, err := cmd.Int64()
	if err != nil {
		err = errors.Wrapf(err, "fail to parse int64 with key %v", key)
		log.Println(err)
		return 0, err
	}
	return value, nil
}

// SetInt64 implements Cache
func (r *RedisCache) SetInt64(ctx context.Context, key string, value int64, d time.Duration) error {
	err := r.rdb.Set(ctx, key, value, d).Err()
	err = errors.Wrapf(err, "fail to set int64 with key %v", key)
	if err != nil {
		log.Printf("cache fail to set int64, err=%v", err)
	}
	return err
}

// TOSAdd implements TimeOrderedSetCache in Redis.
//
// We implement it by maintaining a HASH and a ZSET in Redis:
// - HASH: map tosValue into tosTime, used to ensure the uniqueness of tosValue.
// - ZSET: store tosTime_tosValue items, the order of tosTimeInUnixMs_tosValue
//         is as we expect (first time, second key).
func (r *RedisCache) TOSAdd(ctx context.Context, name string, item TOSItem, maxSize int) error {
	if name == "" || maxSize <= 0 {
		return fmt.Errorf("wrong argument, name %q should be non empty and maxSize %v should be positive", name, maxSize)
	}
	const excludeCharsInName = "{}"
	if strings.ContainsAny(name, excludeCharsInName) {
		return fmt.Errorf("TOS name should exclude charactors in %q", excludeCharsInName)
	}

	// - keys: zsetName, hashName
	//    - zSetName: {name}.zset
	//    - hashName: {name}.hash
	// all keys are with the same hash tag {tos.name} as suffix
	// to make sure all keys are mapped into the same shards,
	// so it's safe to use this script in Redis Cluster.
	//
	// - args: max_size, t, value
	const luaScript = `
        local zSetName = KEYS[1]
        local hashName = KEYS[2]
        local maxSize = tonumber(ARGV[1])
        local tosTime = ARGV[2]
        local tosValue = ARGV[3]
        local zSetMember = tosTime..'_'..tosValue
        
        -- if current size >= max size, remove one
        local currSize = redis.call('ZCARD', zSetName)
        if( currSize >= maxSize ) then
            local maxItem = redis.call('ZPOPMAX', zSetName)
            local removedZSetMember = maxItem[1]
            -- split it according '_' to get the second part, i.e. tosValue
            local removedTOSValue = string.sub(removedSetMember, 1 + string.find(removedZSetMember, '_'))
            redis.call('HREM', hashName, removedTOSValue)
        end

        -- remove old one if exist
        local exist = redis.call('HEXISTS', hashName, tosValue)
        if( exist > 0 ) then
            local oldTOSTime = redis.call('HGET', hashName, tosValue)
            redis.call('ZREM', zSetName, oldTOSTime..'_'..tosValue)
        end

        redis.call('HSET', hashName, tosValue, zSetMember)
        redis.call('ZADD', zSetName, 0, zSetMember)
        return currSize + 1
    `
	zSetName := fmt.Sprintf("{%v}.zset", name)
	hashName := fmt.Sprintf("{%v}.hash", name)
	t := funcs.FullLeadingZeroItoa(item.T.UnixMilli())
	currSize, err := r.rdb.Eval(ctx, luaScript, []string{zSetName, hashName}, maxSize, t, item.Value).Result()
	if err != nil {
		log.Printf("TOSAdd %v fails: %v\n", name, err)
		return errors.Wrapf(err, "fail to execute lua script")
	}

	log.Printf("TOSAdd %v got currSize = %v\n", name, currSize)
	return nil
}

// TOSFetch implements TimeOrderedSetCache
// See TOSAdd(...)
func (r *RedisCache) TOSFetch(ctx context.Context, name string, tCursor time.Time, valueCursor string, n int) ([]TOSItem, error) {
	nMin, nMax := 0, 1000
	if n < nMin || n > nMax {
		return nil, fmt.Errorf("n %v exceed range [%v, %v]", n, nMin, nMax)
	}

	// 1. fetch members from zset, member format: tosTimeInUnixMs_tosValue
	t := funcs.FullLeadingZeroItoa(tCursor.UnixMilli())
	zSetName := fmt.Sprintf("{%v}.zset", name)
	zSetMemberCursor := fmt.Sprintf("(%v_%v", t, valueCursor) // exclusive
	log.Println("Execute ZREVRANGEBYLEX", zSetName, zSetMemberCursor, "-", "LIMIT", 0, n)
	members, err := r.rdb.ZRevRangeByLex(ctx, zSetName, &redis.ZRangeBy{
		Max:    zSetMemberCursor,
		Min:    "-",
		Offset: 0,
		Count:  int64(n),
	}).Result()
	// log.Println("Got members", members)
	if err != nil {
		log.Printf("TOSFetch %v fails: %v\n", name, err)
		return nil, errors.Wrapf(err, "fail to zreverangebylex")
	}

	// 2. parse member
	items := make([]TOSItem, 0, len(members))
	for _, member := range members {
		parts := strings.SplitN(member, "_", 2)
		if len(parts) != 2 {
			err := errors.Errorf("the Redis zset %v member format is wrong, expect part1_part2, got %v\n", zSetName, member)
			log.Println(err)
			return nil, err
		}
		tosTimeStr, tosValue := parts[0], parts[1]
		tosTimeMs, err := strconv.ParseInt(tosTimeStr, 10, 64)
		if err != nil {
			err = errors.Errorf("the Redis zset %v member format is wrong, expect part1 is int64, got %v\n", zSetName, tosTimeStr)
			log.Println(err)
			return nil, err
		}
		items = append(items, TOSItem{
			T:     time.UnixMilli(tosTimeMs),
			Value: tosValue,
		})
	}
	return items, nil
}
