// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func httpHandleRepoRaw(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var rawPathSpec, pathSpec string
	var repo *git.Repository
	var refHash plumbing.Hash
	var refHashSlice []byte
	var commitObj *object.Commit
	var tree *object.Tree
	var err error

	rawPathSpec = params["rest"].(string)
	repo, pathSpec = params["repo"].(*git.Repository), strings.TrimSuffix(rawPathSpec, "/")
	params["path_spec"] = pathSpec

	if refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		errorPage500(writer, params, "Error getting ref hash: "+err.Error())
		return
	}
	refHashSlice = refHash[:]

	cacheHandle := append(refHashSlice, []byte(pathSpec)...)

	if value, found := treeReadmeCache.Get(cacheHandle); found {
		params["files"] = value.DisplayTree
		renderTemplate(writer, "repo_raw_dir", params)
		return
	}
	if value, found := commitPathFileRawCache.Get(cacheHandle); found {
		fmt.Fprint(writer, value)
		return
	}

	if commitObj, err = repo.CommitObject(refHash); err != nil {
		errorPage500(writer, params, "Error getting commit object: "+err.Error())
		return
	}
	if tree, err = commitObj.Tree(); err != nil {
		errorPage500(writer, params, "Error getting file tree: "+err.Error())
		return
	}

	start := time.Now()
	var target *object.Tree
	if pathSpec == "" {
		target = tree
	} else {
		if target, err = tree.Tree(pathSpec); err != nil {
			var file *object.File
			var fileContent string
			if file, err = tree.File(pathSpec); err != nil {
				errorPage500(writer, params, "Error retrieving path: "+err.Error())
				return
			}
			if redirectNoDir(writer, request) {
				return
			}
			if fileContent, err = file.Contents(); err != nil {
				errorPage500(writer, params, "Error reading file: "+err.Error())
				return
			}
			cost := time.Since(start).Nanoseconds()
			commitPathFileRawCache.Set(cacheHandle, fileContent, cost)
			fmt.Fprint(writer, fileContent)
			return
		}
	}

	if redirectDir(writer, request) {
		return
	}

	displayTree := makeDisplayTree(target)
	readmeFilename, readmeRendered := renderReadmeAtTree(target)
	cost := time.Since(start).Nanoseconds()

	params["files"] = displayTree
	params["readme_filename"] = readmeFilename
	params["readme"] = readmeRendered

	treeReadmeCache.Set(cacheHandle, treeReadmeCacheEntry{
		DisplayTree:    displayTree,
		ReadmeFilename: readmeFilename,
		ReadmeRendered: readmeRendered,
	}, cost)

	renderTemplate(writer, "repo_raw_dir", params)
}
