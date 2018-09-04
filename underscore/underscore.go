package underscore

import (
	"strings"
)

func Camel(str string) string {
	var sb strings.Builder
	for _, ch := range str {
		if ch >= 'A' && ch <= 'Z' {
			sb.WriteRune('_')
		}
		sb.WriteRune(ch)
	}
	return strings.ToLower(sb.String())
}
