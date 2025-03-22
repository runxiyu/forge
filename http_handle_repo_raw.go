// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func httpHandleRepoRaw(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var rawPathSpec, pathSpec string
	var repo *git.Repository
	var refHash plumbing.Hash
	var commitObj *object.Commit
	var tree *object.Tree
	var err error

	rawPathSpec = params["rest"].(string)
	repo, pathSpec = params["repo"].(*git.Repository), strings.TrimSuffix(rawPathSpec, "/")
	params["path_spec"] = pathSpec

	if refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		http.Error(writer, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if commitObj, err = repo.CommitObject(refHash); err != nil {
		http.Error(writer, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tree, err = commitObj.Tree(); err != nil {
		http.Error(writer, "Error getting file tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var target *object.Tree
	if pathSpec == "" {
		target = tree
	} else {
		if target, err = tree.Tree(pathSpec); err != nil {
			var file *object.File
			var fileContent string
			if file, err = tree.File(pathSpec); err != nil {
				http.Error(writer, "Error retrieving path: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if redirectNoDir(writer, request) {
				return
			}
			if fileContent, err = file.Contents(); err != nil {
				http.Error(writer, "Error reading file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprint(writer, fileContent)
			return
		}
	}

	if redirectDir(writer, request) {
		return
	}

	params["files"] = makeDisplayTree(target)

	renderTemplate(writer, "repo_raw_dir", params)
}
