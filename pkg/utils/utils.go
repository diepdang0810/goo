package utils

import (
	"strconv"
)

func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func EmptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
