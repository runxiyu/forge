// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// TODO: Make the C part report detailed error messages too

package git2c

import "errors"

var (
	ErrUnknown                         = errors.New("git2c: unknown error")
	ErrPath                            = errors.New("git2c: get tree entry by path failed")
	ErrRevparse                        = errors.New("git2c: revparse failed")
	ErrReadme                          = errors.New("git2c: no readme")
	ErrBlobExpected                    = errors.New("git2c: blob expected")
	ErrEntryToObject                   = errors.New("git2c: tree entry to object conversion failed")
	ErrBlobRawContent                  = errors.New("git2c: get blob raw content failed")
	ErrRevwalk                         = errors.New("git2c: revwalk failed")
	ErrRevwalkPushHead                 = errors.New("git2c: revwalk push head failed")
	ErrBareProto                       = errors.New("git2c: bare protocol error")
	ErrRefResolve                      = errors.New("git2c: ref resolve failed")
	ErrBranches                        = errors.New("git2c: list branches failed")
	ErrCommitLookup                    = errors.New("git2c: commit lookup failed")
	ErrDiff                            = errors.New("git2c: diff failed")
	ErrMergeBaseNone                   = errors.New("git2c: no merge base found")
	ErrMergeBase                       = errors.New("git2c: merge base failed")
	ErrCommitCreate                    = errors.New("git2c: commit create failed")
	ErrUpdateRef                       = errors.New("git2c: update ref failed")
	ErrCommitTree                      = errors.New("git2c: commit tree lookup failed")
	ErrInitRepoCreate                  = errors.New("git2c: init repo: create failed")
	ErrInitRepoConfig                  = errors.New("git2c: init repo: open config failed")
	ErrInitRepoSetHooksPath            = errors.New("git2c: init repo: set core.hooksPath failed")
	ErrInitRepoSetAdvertisePushOptions = errors.New("git2c: init repo: set receive.advertisePushOptions failed")
	ErrInitRepoMkdir                   = errors.New("git2c: init repo: create directory failed")
)

func Perror(errno uint64) error {
	switch errno {
	case 0:
		return nil
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
	case 12:
		return ErrRefResolve
	case 13:
		return ErrBranches
	case 14:
		return ErrCommitLookup
	case 15:
		return ErrDiff
	case 16:
		return ErrMergeBaseNone
	case 17:
		return ErrMergeBase
	case 18:
		return ErrUpdateRef
	case 19:
		return ErrCommitCreate
	case 20:
		return ErrInitRepoCreate
	case 21:
		return ErrInitRepoConfig
	case 22:
		return ErrInitRepoSetHooksPath
	case 23:
		return ErrInitRepoSetAdvertisePushOptions
	case 24:
		return ErrInitRepoMkdir
	}
	return ErrUnknown
}
