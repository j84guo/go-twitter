package main

import (
	"fmt"
)

func shouldEncode(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}
	switch c {
	case '-', '.', '_', '~':
		return false
	default:
		return true
	}
}

func PercentEncode(str string) string {
	CHARS := "0123456789ABCDEF"

	hex := 0
	for i := range str {
		if shouldEncode(str[i]) {
			hex++
		}
	}

	buf := make([]byte, len(str) + 2 * hex)
	j := 0
	for i := range str {
		if shouldEncode(str[i]) {
			buf[j] = '%'
			buf[j + 1] = CHARS[str[i] >> 4]
			buf[j + 2] = CHARS[str[i] & 0xF]
			j += 2
		} else {
			buf[j] = str[i]
		}
		j++
	}

	return string(buf)
}

func main() {
	fmt.Println("Percent Encoding")
	fmt.Println(PercentEncode("ella "))
}
