// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

import "iter"

// iterSeqLimit returns an iterator equivalent to the supplied one, but stops
// after n iterations.
func IterSeqLimit[T any](s iter.Seq[T], n uint) iter.Seq[T] {
	return func(yield func(T) bool) {
		var iterations uint
		for v := range s {
			if iterations > n-1 {
				return
			}
			if !yield(v) {
				return
			}
			iterations++
		}
	}
}
