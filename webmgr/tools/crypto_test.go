package tools

import (
	"errors"
	"testing"
)

func TestPassword(t *testing.T) {
	h, err := HashPassword("admin")
	if err != nil {
		t.Error(err)
	}
	t.Log(h)
	if !ValidatePassword(h, "admin") {
		t.Error(errors.New("password mismatched"))
	}
}
