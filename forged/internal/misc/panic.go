// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

func FirstOrPanic[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func NoneOrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
