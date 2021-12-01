package util

import (
	"strconv"
)

func ParseInt(s string) (int, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}
