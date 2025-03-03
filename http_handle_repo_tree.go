// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path"
	"strings"

	chroma_formatters_html "github.com/alecthomas/chroma/v2/formatters/html"
	chroma_lexers "github.com/alecthomas/chroma/v2/lexers"
	chroma_styles "github.com/alecthomas/chroma/v2/styles"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_tree(w http.ResponseWriter, r *http.Request, params map[string]any) {
	raw_path_spec := params["rest"].(string)
	repo, path_spec := params["repo"].(*git.Repository), strings.TrimSuffix(raw_path_spec, "/")
	params["path_spec"] = path_spec

	ref_hash, err := get_ref_hash_from_type_and_name(repo, params["ref_type"].(string), params["ref_name"].(string))
	if err != nil {
		http.Error(w, "Error getting ref hash: "+err.Error(), http.StatusInternalServerError)
		return
	}
	commit_object, err := repo.CommitObject(ref_hash)
	if err != nil {
		http.Error(w, "Error getting commit object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		http.Error(w, "Error getting file tree: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var target *object.Tree
	if path_spec == "" {
		target = tree
	} else {
		target, err = tree.Tree(path_spec)
		if err != nil {
			file, err := tree.File(path_spec)
			if err != nil {
				http.Error(w, "Error retrieving path: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] == '/' {
				http.Redirect(w, r, "../"+path_spec, http.StatusSeeOther)
				return
			}
			file_contents, err := file.Contents()
			if err != nil {
				http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			lexer := chroma_lexers.Match(path_spec)
			if lexer == nil {
				lexer = chroma_lexers.Fallback
			}
			iterator, err := lexer.Tokenise(nil, file_contents)
			if err != nil {
				http.Error(w, "Error tokenizing code: "+err.Error(), http.StatusInternalServerError)
				return
			}
			var formatted_unencapsulated bytes.Buffer
			style := chroma_styles.Get("autumn")
			formatter := chroma_formatters_html.New(chroma_formatters_html.WithClasses(true), chroma_formatters_html.TabWidth(8))
			err = formatter.Format(&formatted_unencapsulated, style, iterator)
			if err != nil {
				http.Error(w, "Error formatting code: "+err.Error(), http.StatusInternalServerError)
				return
			}
			formatted_encapsulated := template.HTML(formatted_unencapsulated.Bytes())
			params["file_contents"] = formatted_encapsulated

			render_template(w, "repo_tree_file", params)
			return
		}
	}

	if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] != '/' {
		http.Redirect(w, r, path.Base(path_spec)+"/", http.StatusSeeOther)
		return
	}

	params["readme_filename"], params["readme"] = render_readme_at_tree(target)
	params["files"] = build_display_git_tree(target)

	render_template(w, "repo_tree_dir", params)
}
