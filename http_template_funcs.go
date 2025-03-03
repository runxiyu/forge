// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"path"
	"strings"
)

func first_line(s string) string {
	before, _, _ := strings.Cut(s, "\n")
	return before
}

func base_name(s string) string {
	return path.Base(s)
}
