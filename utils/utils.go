package utils

import "strconv"

func StrMustInt(s string) int {
	i, _ := strconv.ParseInt(s, 10, 64)
	return int(i)
}
