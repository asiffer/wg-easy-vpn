package utils

import "strings"

var (
	// CommentPrefixes are the strings which start a comment
	CommentPrefixes = []string{"#", ";", "//"}
)

func removeComment(line string) string {
	for _, pre := range CommentPrefixes {
		if strings.HasPrefix(line, pre) {
			return ""
		}
	}
	return line
}
