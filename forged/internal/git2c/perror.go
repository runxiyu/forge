// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// TODO: Make the C part report detailed error messages too

package git2c

import "errors"

var (
	Success            error
	ErrUnknown         = errors.New("git2c: unknown error")
	ErrPath            = errors.New("git2c: get tree entry by path failed")
	ErrRevparse        = errors.New("git2c: revparse failed")
	ErrReadme          = errors.New("git2c: no readme")
	ErrBlobExpected    = errors.New("git2c: blob expected")
	ErrEntryToObject   = errors.New("git2c: tree entry to object conversion failed")
	ErrBlobRawContent  = errors.New("git2c: get blob raw content failed")
	ErrRevwalk         = errors.New("git2c: revwalk failed")
	ErrRevwalkPushHead = errors.New("git2c: revwalk push head failed")
	ErrBareProto       = errors.New("git2c: bare protocol error")
)

func Perror(errno uint) error {
	switch errno {
	case 0:
		return Success
	case 3:
		return ErrPath
	case 4:
		return ErrRevparse
	case 5:
		return ErrReadme
	case 6:
		return ErrBlobExpected
	case 7:
		return ErrEntryToObject
	case 8:
		return ErrBlobRawContent
	case 9:
		return ErrRevwalk
	case 10:
		return ErrRevwalkPushHead
	case 11:
		return ErrBareProto
	}
	return ErrUnknown
}
