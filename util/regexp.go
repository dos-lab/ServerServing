package util

import "regexp"

func SplitLine(s string) []string {
	splitter := regexp.MustCompile(`\r\n|\r|\n`)
	lines := splitter.Split(s, -1)
	return lines
}

func SplitSpaces(s string) []string {
	splitter := regexp.MustCompile(`\s+`)
	lines := splitter.Split(s, -1)
	return lines
}