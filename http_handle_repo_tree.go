package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"

	chroma_formatters_html "github.com/alecthomas/chroma/v2/formatters/html"
	chroma_lexers "github.com/alecthomas/chroma/v2/lexers"
	chroma_styles "github.com/alecthomas/chroma/v2/styles"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func handle_repo_tree(w http.ResponseWriter, r *http.Request, params map[string]any) {
	raw_path_spec := params["rest"].(string)
	group_name, repo_name, path_spec := params["group_name"].(string), params["repo_name"].(string), strings.TrimSuffix(raw_path_spec, "/")
	ref_type, ref_name, err := get_param_ref_and_type(r)
	if err != nil {
		if errors.Is(err, err_no_ref_spec) {
			ref_type = "head"
		} else {
			fmt.Fprintln(w, "Error querying ref type:", err.Error())
			return
		}
	}
	params["ref_type"], params["ref"], params["path_spec"] = ref_type, ref_name, path_spec
	repo, description, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		fmt.Fprintln(w, "Error opening repo:", err.Error())
		return
	}
	params["repo_description"] = description

	ref_hash, err := get_ref_hash_from_type_and_name(repo, ref_type, ref_name)
	if err != nil {
		fmt.Fprintln(w, "Error getting ref hash:", err.Error())
		return
	}
	commit_object, err := repo.CommitObject(ref_hash)
	if err != nil {
		fmt.Fprintln(w, "Error getting commit object:", err.Error())
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		fmt.Fprintln(w, "Error getting file tree:", err.Error())
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
				fmt.Fprintln(w, "Error retrieving path:", err.Error())
				return
			}
			if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] == '/' {
				http.Redirect(w, r, "../"+path_spec, http.StatusSeeOther)
				return
			}
			file_contents, err := file.Contents()
			if err != nil {
				fmt.Fprintln(w, "Error reading file:", err.Error())
				return
			}
			lexer := chroma_lexers.Match(path_spec)
			if lexer == nil {
				lexer = chroma_lexers.Fallback
			}
			iterator, err := lexer.Tokenise(nil, file_contents)
			if err != nil {
				fmt.Fprintln(w, "Error tokenizing code:", err.Error())
				return
			}
			var formatted_unencapsulated bytes.Buffer
			style := chroma_styles.Get("autumn")
			formatter := chroma_formatters_html.New(chroma_formatters_html.WithClasses(true), chroma_formatters_html.TabWidth(8))
			err = formatter.Format(&formatted_unencapsulated, style, iterator)
			if err != nil {
				fmt.Fprintln(w, "Error formatting code:", err.Error())
				return
			}
			formatted_encapsulated := template.HTML(formatted_unencapsulated.Bytes())
			params["file_contents"] = formatted_encapsulated

			err = templates.ExecuteTemplate(w, "repo_tree_file", params)
			if err != nil {
				fmt.Fprintln(w, "Error rendering template:", err.Error())
				return
			}
			return
		}
	}

	if len(raw_path_spec) != 0 && raw_path_spec[len(raw_path_spec)-1] != '/' {
		http.Redirect(w, r, path.Base(path_spec)+"/", http.StatusSeeOther)
		return
	}

	params["readme_filename"], params["readme"] = render_readme_at_tree(target)
	params["files"] = build_display_git_tree(target)

	err = templates.ExecuteTemplate(w, "repo_tree_dir", params)
	if err != nil {
		fmt.Fprintln(w, "Error rendering template:", err.Error())
		return
	}
}
