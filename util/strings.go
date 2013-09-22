package util

import . "strings"

func Split2(s, sep string) (string, string) {
	if i := Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):]
	}
	return s, ""
}
