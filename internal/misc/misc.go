// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

import "strings"

func FirstOrPanic[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// sliceContainsNewlines returns true if and only if the given slice contains
// one or more strings that contains newlines.
func SliceContainsNewlines(s []string) bool {
	for _, v := range s {
		if strings.Contains(v, "\n") {
			return true
		}
	}
	return false
}
