package matcher

import (
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

func stringsIndexAny(str, chars string) int {
	ls := len(str)
	lc := len(chars)

	for i := 0; i < ls; i++ {
		s := str[i]
		for j := 0; j < lc; j++ {
			if s == chars[j] {
				return i
			}
		}
	}
	return -1
}

func stringsIndexAnyNotEscaped(str, chars string) int {
	ls := len(str)
	lc := len(chars)

	for i := 0; i < ls; i++ {
		s := str[i]
		for j := 0; j < lc; j++ {
			if s == chars[j] {
				if i > 0 && str[i-1] == '\\' {
					continue
				}
				return i
			}
		}
	}
	return -1
}

func stringsIndex(str string, char byte) int {
	ls := len(str)

	for i := 0; i < ls; i++ {
		if str[i] == char {
			return i
		}
	}
	return -1
}

func stringsHasPrefix(str, prefix string) bool {
	// ls := len(str)
	sl := len(str)
	pl := len(prefix)
	if sl < pl {
		return false
	}
	i := 0
	for ; i < pl; i++ {
		if str[i] != prefix[i] {
			break
		}
	}
	if i == pl {
		return true
	}
	return false
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
	for i := 0; i < len(chars); i++ {
		if label == chars[i] {
			return true
		}
	}
	return false
}

func isInStringSlice(slice []string, expected string) bool {
	for _, val := range slice {
		if val == expected {
			return true
		}
	}
	return false
}

// nextParenthesis finds the starting and ending indices of the openning and closing parenthesis characters: '(' and ')'
// inspired by https://github.com/gorilla/mux/blob/master/regexp.go#L214
func nextParenthesis(pattern string) (start, end int) {
	level := 0
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch c {
		case '(':
			level++
			if level == 1 {
				start = i
			}
		case ')':
			level--
			if level == 0 {
				end = i
				return
			} else if level < 0 {
				panicm("too many closed parenthesis in %s", pattern)
			}
		}
	}

	if level != 0 {
		panicm("unbalanced parenthesis in %s", pattern)
	}

	return
}
