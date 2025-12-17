package encode

import (
	"crypto/md5"
	"encoding/hex"
	"hash/crc32"
	"strconv"
)

// CalMd5 calculates MD5 of bytes.
func CalMd5(b []byte) string {
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:])
}

// Crc32HashCode produces a non-negative CRC32 hash string.
func Crc32HashCode(b []byte) string {
	v := int(crc32.ChecksumIEEE(b))
	if v >= 0 {
		return strconv.Itoa(v)
	}
	if -v >= 0 {
		return strconv.Itoa(-v)
	}
	return ""
}
