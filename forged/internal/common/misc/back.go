// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

// ErrorBack wraps a value and a channel for communicating an associated error.
// Typically used to get an error response after sending data across a channel.
type ErrorBack[T any] struct {
	Content   T
	ErrorChan chan error
}
