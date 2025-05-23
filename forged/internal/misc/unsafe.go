// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

import "unsafe"

// StringToBytes converts a string to a byte slice without copying the string.
// Memory is borrowed from the string.
// The resulting byte slice must not be modified in any form.
func StringToBytes(s string) (bytes []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts a byte slice to a string without copying the bytes.
// Memory is borrowed from the byte slice.
// The source byte slice must not be modified.
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
