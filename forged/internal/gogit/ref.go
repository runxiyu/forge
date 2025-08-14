// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gogit

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GetRefHash returns the hash of a reference given its type and name.
func GetRefHash(repo *git.Repository, refType, refName string) (plumbing.Hash, error) {
	var ref *plumbing.Reference
	var err error

	switch refType {
	case "":
		if ref, err = repo.Head(); err != nil {
			return plumbing.Hash{}, err
		}
		return ref.Hash(), nil
	case "commit":
		return plumbing.NewHash(refName), nil
	case "branch":
		if ref, err = repo.Reference(plumbing.NewBranchReferenceName(refName), true); err != nil {
			return plumbing.Hash{}, err
		}
		return ref.Hash(), nil
	case "tag":
		if ref, err = repo.Reference(plumbing.NewTagReferenceName(refName), true); err != nil {
			return plumbing.Hash{}, err
		}
		return ref.Hash(), nil
	default:
		panic("invalid ref type " + refType)
	}
}
