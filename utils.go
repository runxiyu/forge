// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import "strings"

func sliceContainsNewlines(s []string) bool {
	for _, v := range s {
		if strings.Contains(v, "\n") {
			return true
		}
	}
	return false
}
