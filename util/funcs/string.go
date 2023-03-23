package funcs

import "strconv"

func Itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Atoi(a string) (int64, error) {
	return strconv.ParseInt(a, 10, 64)
}
