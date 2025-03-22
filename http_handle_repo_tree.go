// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromaHTML "github.com/alecthomas/chroma/v2/formatters/html"
	chromaLexers "github.com/alecthomas/chroma/v2/lexers"
	chromaStyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func httpHandleRepoTree(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var rawPathSpec, pathSpec string
	var repo *git.Repository
	var refHash plumbing.Hash
	var commitObject *object.Commit
	var tree *object.Tree
	var err error

	rawPathSpec = params["rest"].(string)
	repo, pathSpec = params["repo"].(*git.Repository), strings.TrimSuffix(rawPathSpec, "/")
	params["path_spec"] = pathSpec

	if refHash, err = getRefHash(repo, params["ref_type"].(string), params["ref_name"].(string)); err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if commitObject, err = repo.CommitObject(refHash); err != nil {
		http.Error(w, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if tree, err = commitObject.Tree(); err != nil {
		http.Error(w, "Error getting file tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var target *object.Tree
	if pathSpec == "" {
		target = tree
	} else {
		if target, err = tree.Tree(pathSpec); err != nil {
			var file *object.File
			var fileContent string
			var lexer chroma.Lexer
			var iterator chroma.Iterator
			var style *chroma.Style
			var formatter *chromaHTML.Formatter
			var formattedHTML template.HTML

			if file, err = tree.File(pathSpec); err != nil {
				http.Error(w, "Error retrieving path: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if redirectNoDir(w, r) {
				return
			}
			if fileContent, err = file.Contents(); err != nil {
				http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			lexer = chromaLexers.Match(pathSpec)
			if lexer == nil {
				lexer = chromaLexers.Fallback
			}
			if iterator, err = lexer.Tokenise(nil, fileContent); err != nil {
				http.Error(w, "Error tokenizing code: "+err.Error(), http.StatusInternalServerError)
				return
			}
			var formattedHTMLStr bytes.Buffer
			style = chromaStyles.Get("autumn")
			formatter = chromaHTML.New(chromaHTML.WithClasses(true), chromaHTML.TabWidth(8))
			if err = formatter.Format(&formattedHTMLStr, style, iterator); err != nil {
				http.Error(w, "Error formatting code: "+err.Error(), http.StatusInternalServerError)
				return
			}
			formattedHTML = template.HTML(formattedHTMLStr.Bytes()) //#nosec G203
			params["file_contents"] = formattedHTML

			renderTemplate(w, "repo_tree_file", params)
			return
		}
	}

	if len(rawPathSpec) != 0 && rawPathSpec[len(rawPathSpec)-1] != '/' {
		http.Redirect(w, r, path.Base(pathSpec)+"/", http.StatusSeeOther)
		return
	}

	params["readme_filename"], params["readme"] = renderReadmeAtTree(target)
	params["files"] = makeDisplayTree(target)

	renderTemplate(w, "repo_tree_dir", params)
}
