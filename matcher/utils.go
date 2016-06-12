package matcher

import (
	"bytes"
	"fmt"
	"strings"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestPrefix(s1, s2 string) int {
	max := min(len(s1), len(s2))
	i := 0
	for i < max && s1[i] == s2[i] {
		i++
	}
	return i
}

func panicm(format string, args ...interface{}) {
	panic(fmt.Sprintf("lion: "+format, args...))
}

func reverseHost(pattern string) string {
	reversed := strings.Split(pattern, ".")
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	return strings.Join(reversed, ".")
}

func isByteInString(label byte, chars string) bool {
	return bytes.IndexAny([]byte{label}, chars) != -1
}

func isInStringSlice(slice []string, expected string) bool {
	for _, val := range slice {
		if val == expected {
			return true
		}
	}
	return false
}
