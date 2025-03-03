// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// get_ref_hash_from_type_and_name returns the hash of a reference given its
// type and name as supplied in URL queries.
func get_ref_hash_from_type_and_name(repo *git.Repository, ref_type, ref_name string) (ref_hash plumbing.Hash, ret_err error) {
	switch ref_type {
	case "":
		head, err := repo.Head()
		if err != nil {
			ret_err = err
			return
		}
		ref_hash = head.Hash()
	case "commit":
		ref_hash = plumbing.NewHash(ref_name)
	case "branch":
		ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
		if err != nil {
			ret_err = err
			return
		}
		ref_hash = ref.Hash()
	case "tag":
		ref, err := repo.Reference(plumbing.NewTagReferenceName(ref_name), true)
		if err != nil {
			ret_err = err
			return
		}
		ref_hash = ref.Hash()
	default:
		panic("Invalid ref type " + ref_type)
	}
	return
}
