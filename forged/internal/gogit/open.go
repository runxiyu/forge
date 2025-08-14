// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gogit

import "github.com/go-git/go-git/v5"

// Repository is an alias of [git.Repository].
type Repository = git.Repository

// Open opens a git repository at the given path.
func Open(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}
