// cache keys compose functions
// format: PROJECT:part1:part2:...
package keys

import (
	"fmt"
	"strings"
)

const (
	PROJECT = "hoyobar"
)

func AuthToken(token string) string {
	return Key("auth_token", token)
}

func Key(parts ...interface{}) string {
	strs := make([]string, 1+len(parts))
	strs[0] = PROJECT
	for i, part := range parts {
		strs[i+1] = fmt.Sprintf("%v", part)
	}
	return strings.Join(strs, ":")
}

func UserBasic(userID int64) string {
	return Key("user", userID, "basic")
}

func EmailToUserID(email string) string {
	return Key("email", email, "user_id")
}

func PhoneToUserID(phone string) string {
	return Key("phone", phone, "user_id")
}

func NicknameToUserID(nickname string) string {
	return Key("nickname", nickname, "user_id")
}

func PostBasic(postID int64) string {
	return Key("post", postID, "basic")
}

func PostContent(postID int64) string {
	return Key("post", postID, "content")
}

func PostReplyNum(postID int64) string {
	return Key("post", postID, "reply_num")
}

func PostReplyTime(postID int64) string {
	return Key("post", postID, "reply_time")
}

func PostListName(order string) string {
	return Key("post", "post_list_order_by_"+order)
}
