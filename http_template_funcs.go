// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"net/url"
	"path"
	"strings"
)

func firstLine(s string) string {
	before, _, _ := strings.Cut(s, "\n")
	return before
}

func baseName(s string) string {
	return path.Base(s)
}

func pathEscape(s string) string {
	return url.PathEscape(s)
}

func queryEscape(s string) string {
	return url.QueryEscape(s)
}

func dereference[T any](p *T) T {
	return *p
}

func dereferenceOrZero[T any](p *T) T {
	if p != nil {
		return *p
	}
	var z T
	return z
}

func minus(a, b int) int {
	return a - b
}
