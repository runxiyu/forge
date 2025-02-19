package main

import (
	"errors"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

var (
	err_getting_tag_reference    = errors.New("error getting tag reference")
	err_getting_branch_reference = errors.New("error getting branch reference")
	err_getting_head             = errors.New("error getting HEAD")
)

// get_ref_hash_from_type_and_name returns the hash of a reference given its
// type and name as supplied in URL queries.
func get_ref_hash_from_type_and_name(repo *git.Repository, ref_type, ref_name string) (ref_hash plumbing.Hash, ret_err error) {
	switch ref_type {
	case "":
		head, err := repo.Head()
		if err != nil {
			ret_err = misc.Wrap_one_error(err_getting_head, err)
			return
		}
		ref_hash = head.Hash()
	case "commit":
		ref_hash = plumbing.NewHash(ref_name)
	case "branch":
		ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
		if err != nil {
			ret_err = misc.Wrap_one_error(err_getting_branch_reference, err)
			return
		}
		ref_hash = ref.Hash()
	case "tag":
		ref, err := repo.Reference(plumbing.NewTagReferenceName(ref_name), true)
		if err != nil {
			ret_err = misc.Wrap_one_error(err_getting_tag_reference, err)
			return
		}
		ref_hash = ref.Hash()
	default:
		panic("Invalid ref type " + ref_type)
	}
	return
}
