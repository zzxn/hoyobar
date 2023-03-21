package myhash

import (
	"hash/fnv"
	"math"
)

// mod should > 0
// non-cryptographic
func HashString(s string, mod int64) int64 {
	h := fnv.New64()
	h.Write([]byte(s))               // ignore err, impossible to fail
	res := h.Sum64() & math.MaxInt64 // ignore the sign bit
	return int64(res) % mod
}

const (
	move0 = 8 * iota
	move1
	move2
	move3
	move4
	move5
	move6
	move7
)

// mod should > 0
// non-cryptographic
func HashSnowflakeID(i int64, mod int64) int64 {
	// necessary: mix high bits, 23 is for 2ms position for snowflake id
	// we use 23 instead of 22 is because it seems on some machine, time in ms tend to be even
	i ^= (i >> 23)

	h := fnv.New64()
	h.Write([]byte{
		byte((i >> move0) & 0xff),
		byte((i >> move1) & 0xff),
		byte((i >> move2) & 0xff),
		byte((i >> move3) & 0xff),
		byte((i >> move4) & 0xff),
		byte((i >> move5) & 0xff),
		byte((i >> move6) & 0xff),
		byte((i >> move7) & 0xff),
	})
	return i % mod
}
