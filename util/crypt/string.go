package crypt

import "hash/adler32"

func HashString(s string, mod int64) int64 {
	var cs uint32 = adler32.Checksum([]byte(s))
	h := int64(cs) % mod // must be positive
	return h
}
