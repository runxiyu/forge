// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	chromaHTML "github.com/alecthomas/chroma/v2/formatters/html"
	chromaLexers "github.com/alecthomas/chroma/v2/lexers"
	chromaStyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func httpHandleRepoTree(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var rawPathSpec, pathSpec string
	var repo *git.Repository
	var refHash plumbing.Hash
	var refHashSlice []byte
	var commitObject *object.Commit
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

	var target *object.Tree
	if pathSpec == "" {
		if value, found := treeReadmeCache.Get(refHashSlice); found {
			params["files"] = value.DisplayTree
			params["readme_filename"] = value.ReadmeFilename
			params["readme"] = value.ReadmeRendered
		} else {
			if commitObject, err = repo.CommitObject(refHash); err != nil {
				errorPage500(writer, params, "Error getting commit object: "+err.Error())
				return
			}
			if tree, err = commitObject.Tree(); err != nil {
				errorPage500(writer, params, "Error getting file tree: "+err.Error())
				return
			}

			start := time.Now()
			displayTree := makeDisplayTree(tree)
			readmeFilename, readmeRendered := renderReadmeAtTree(tree)
			cost := time.Since(start).Nanoseconds()

			params["files"] = displayTree
			params["readme_filename"] = readmeFilename
			params["readme"] = readmeRendered

			entry := treeReadmeCacheEntry{
				DisplayTree:    displayTree,
				ReadmeFilename: readmeFilename,
				ReadmeRendered: readmeRendered,
			}
			treeReadmeCache.Set(refHashSlice, entry, cost)
		}

		renderTemplate(writer, "repo_tree_dir", params)
		return
	}

	if commitObject, err = repo.CommitObject(refHash); err != nil {
		errorPage500(writer, params, "Error getting commit object: "+err.Error())
		return
	}
	if tree, err = commitObject.Tree(); err != nil {
		errorPage500(writer, params, "Error getting file tree: "+err.Error())
		return
	}
	if target, err = tree.Tree(pathSpec); err != nil {
		var file *object.File
		var fileContent string
		var lexer chroma.Lexer
		var iterator chroma.Iterator
		var style *chroma.Style
		var formatter *chromaHTML.Formatter
		var formattedHTML template.HTML

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
		lexer = chromaLexers.Match(pathSpec)
		if lexer == nil {
			lexer = chromaLexers.Fallback
		}
		if iterator, err = lexer.Tokenise(nil, fileContent); err != nil {
			errorPage500(writer, params, "Error tokenizing code: "+err.Error())
			return
		}
		var formattedHTMLStr bytes.Buffer
		style = chromaStyles.Get("autumn")
		formatter = chromaHTML.New(chromaHTML.WithClasses(true), chromaHTML.TabWidth(8))
		if err = formatter.Format(&formattedHTMLStr, style, iterator); err != nil {
			errorPage500(writer, params, "Error formatting code: "+err.Error())
			return
		}
		formattedHTML = template.HTML(formattedHTMLStr.Bytes()) //#nosec G203
		params["file_contents"] = formattedHTML

		renderTemplate(writer, "repo_tree_file", params)
		return
	}

	if len(rawPathSpec) != 0 && rawPathSpec[len(rawPathSpec)-1] != '/' {
		http.Redirect(writer, request, path.Base(pathSpec)+"/", http.StatusSeeOther)
		return
	}

	params["readme_filename"], params["readme"] = renderReadmeAtTree(target)
	params["files"] = makeDisplayTree(target)

	renderTemplate(writer, "repo_tree_dir", params)
}
