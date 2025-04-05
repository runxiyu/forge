// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"net/url"
	"strings"
)

// These are all trivial functions that are intended to be used in HTML
// templates.

// FirstLine returns the first line of a string.
func FirstLine(s string) string {
	before, _, _ := strings.Cut(s, "\n")
	return before
}

// PathEscape escapes the input as an URL path segment.
func PathEscape(s string) string {
	return url.PathEscape(s)
}

// QueryEscape escapes the input as an URL query segment.
func QueryEscape(s string) string {
	return url.QueryEscape(s)
}

// Dereference dereferences a pointer.
func Dereference[T any](p *T) T {
	return *p
}

// DereferenceOrZero dereferences a pointer. If the pointer is nil, the zero
// value of its associated type is returned instead.
func DereferenceOrZero[T any](p *T) T {
	if p != nil {
		return *p
	}
	var z T
	return z
}

// Minus subtracts two numbers.
func Minus(a, b int) int {
	return a - b
}
