package util

import "regexp"

func SplitLine(s string) []string {
	splitter := regexp.MustCompile(`\r\n|\r|\n`)
	lines := splitter.Split(s, -1)
	return lines
}
