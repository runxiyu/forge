package main

import (
	"path/filepath"
	"strings"
)

func first_line(s string) string {
	before, _, _ := strings.Cut(s, "\n")
	return before
}

func base_name (s string) string {
	return filepath.Base(s)
}
