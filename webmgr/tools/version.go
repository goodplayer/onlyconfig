package tools

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
)

func VersionToString(n int64) string {
	if n <= 0 {
		panic(errors.New("version number should not be negative:" + strconv.FormatInt(n, 10)))
	}
	v := make([]byte, 8)
	binary.BigEndian.PutUint64(v, uint64(n))
	return "v" + hex.EncodeToString(v)
}
