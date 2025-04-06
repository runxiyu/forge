// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

// FirstOrPanic returns the value or panics if the error is non-nil.
func FirstOrPanic[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// NoneOrPanic panics if the provided error is non-nil.
func NoneOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
