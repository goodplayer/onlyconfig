package tools

import (
	"math"
	"testing"
)

func TestVersionToString(t *testing.T) {
	assertVersionToString(t, 1, "v0000000000000001")
	assertVersionToString(t, 2, "v0000000000000002")
	assertVersionToString(t, 100, "v0000000000000064")
	assertVersionToString(t, math.MaxInt64, "v7fffffffffffffff")
}

func assertVersionToString(t *testing.T, n int64, expected string) {
	if VersionToString(n) != expected {
		t.Fatal("unexpected value:"+VersionToString(n), "expected:"+expected)
	}
}
