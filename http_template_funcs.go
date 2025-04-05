// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"net/url"
	"strings"
)

// These are all trivial functions that are used in HTML templates.
// See resources.go.

// firstLine returns the first line of a string.
func firstLine(s string) string {
	before, _, _ := strings.Cut(s, "\n")
	return before
}

// pathEscape escapes the input as an URL path segment.
func pathEscape(s string) string {
	return url.PathEscape(s)
}

// queryEscape escapes the input as an URL query segment.
func queryEscape(s string) string {
	return url.QueryEscape(s)
}

// dereference dereferences a pointer.
func dereference[T any](p *T) T {
	return *p
}

// dereferenceOrZero dereferences a pointer. If the pointer is nil, the zero
// value of its associated type is returned instead.
func dereferenceOrZero[T any](p *T) T {
	if p != nil {
		return *p
	}
	var z T
	return z
}

// minus subtracts two numbers.
func minus(a, b int) int {
	return a - b
}
