// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gogit

import (
	"context"
	"errors"
	"io"
	"iter"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitIterSeqErr creates an [iter.Seq[*object.Commit]] from an
// [object.CommitIter], and additionally returns a pointer to error.
// The pointer to error is guaranteed to be populated with either nil or the
// error returned by the commit iterator after the returned iterator is
// finished.
func CommitIterSeqErr(ctx context.Context, commitIter object.CommitIter) (iter.Seq[*object.Commit], *error) {
	var err error
	return func(yield func(*object.Commit) bool) {
		for {
			commit, err2 := commitIter.Next()
			if err2 != nil {
				if errors.Is(err2, io.EOF) {
					return
				}
				err = err2
				return
			}

			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			default:
			}

			if !yield(commit) {
				return
			}
		}
	}, &err
}
