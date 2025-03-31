// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// getRefHash returns the hash of a reference given its
// type and name as supplied in URL queries.
func getRefHash(repo *git.Repository, refType, refName string) (refHash plumbing.Hash, err error) {
	var ref *plumbing.Reference
	switch refType {
	case "":
		if ref, err = repo.Head(); err != nil {
			return
		}
		refHash = ref.Hash()
	case "commit":
		refHash = plumbing.NewHash(refName)
	case "branch":
		if ref, err = repo.Reference(plumbing.NewBranchReferenceName(refName), true); err != nil {
			return
		}
		refHash = ref.Hash()
	case "tag":
		if ref, err = repo.Reference(plumbing.NewTagReferenceName(refName), true); err != nil {
			return
		}
		refHash = ref.Hash()
	default:
		panic("Invalid ref type " + refType)
	}
	return
}
