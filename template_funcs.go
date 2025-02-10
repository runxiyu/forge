package main

import "strings"

func first_line(s string) string {
	before, _, _ := strings.Cut(s, "\n")
	return before
}
