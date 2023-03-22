// cache keys compose functions
package keys

import "fmt"

func AuthToken(token string) string {
	return fmt.Sprintf("hoyobar:auth_token:%v", token)
}
