// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import "strings"

// sliceContainsNewlines returns true if and only if the given slice contains
// one or more strings that contains newlines.
func sliceContainsNewlines(s []string) bool {
	for _, v := range s {
		if strings.Contains(v, "\n") {
			return true
		}
	}
	return false
}
