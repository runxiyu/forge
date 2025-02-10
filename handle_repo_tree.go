package main

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	chroma_formatters_html "github.com/alecthomas/chroma/v2/formatters/html"
	chroma_lexers "github.com/alecthomas/chroma/v2/lexers"
	chroma_styles "github.com/alecthomas/chroma/v2/styles"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

func handle_repo_tree(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	// TODO: Sanitize path values
	ref_name, category_name, repo_name, path_spec := r.PathValue("ref"), r.PathValue("category_name"), r.PathValue("repo_name"), strings.TrimSuffix(r.PathValue("rest"), "/")
	data["category_name"], data["repo_name"], data["path_spec"] = category_name, repo_name, path_spec
	repo, err := git.PlainOpen(filepath.Join(config.Git.Root, category_name, repo_name+".git"))
	if err != nil {
		_, _ = w.Write([]byte("Error opening repo: " + err.Error()))
		return
	}
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
	if err != nil {
		_, _ = w.Write([]byte("Error getting repo reference: " + err.Error()))
		return
	}
	ref_hash := ref.Hash()
	commit_object, err := repo.CommitObject(ref_hash)
	if err != nil {
		_, _ = w.Write([]byte("Error getting commit object: " + err.Error()))
		return
	}
	tree, err := commit_object.Tree()
	if err != nil {
		_, _ = w.Write([]byte("Error getting file tree: " + err.Error()))
		return
	}

	target, err := tree.Tree(path_spec)
	if err != nil {
		file, err := tree.File(path_spec)
		if err != nil {
			_, _ = w.Write([]byte("Error retrieving path: " + err.Error()))
			return
		}
		file_contents, err := file.Contents()
		if err != nil {
			_, _ = w.Write([]byte("Error reading file: " + err.Error()))
			return
		}
		lexer := chroma_lexers.Match(path_spec)
		if lexer == nil {
			lexer = chroma_lexers.Fallback
		}
		iterator, err := lexer.Tokenise(nil, file_contents)
		if err != nil {
			_, _ = w.Write([]byte("Error rendering code: " + err.Error()))
			return
		}
		var formatted_unencapsulated bytes.Buffer
		style := chroma_styles.Get("autumn")
		formatter := chroma_formatters_html.New(chroma_formatters_html.WithClasses(true), chroma_formatters_html.TabWidth(8))
		formatter.Format(&formatted_unencapsulated, style, iterator)
		formatted_encapsulated := template.HTML(formatted_unencapsulated.Bytes())
		data["file_contents"] = formatted_encapsulated

		err = templates.ExecuteTemplate(w, "repo_tree_file", data)
		if err != nil {
			_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
			return
		}
		return
	}

	readme_file, err := target.File("README.md")
	if err != nil {
		data["readme"] = "There is no README for this directory."
	} else {
		readme_file_contents, err := readme_file.Contents()
		if err != nil {
			data["readme"] = "There is no README for this directory."
		} else {
			var readme_rendered_unsafe bytes.Buffer
			err = goldmark.Convert([]byte(readme_file_contents), &readme_rendered_unsafe)
			if err != nil {
				readme_rendered_unsafe.WriteString("Unable to render README: " + err.Error())
			}
			readme_rendered_safe := template.HTML(bluemonday.UGCPolicy().SanitizeBytes(readme_rendered_unsafe.Bytes()))
			data["readme"] = readme_rendered_safe
		}
	}

	display_git_tree := make([]display_git_tree_entry_t, 0)
	for _, entry := range target.Entries {
		display_git_tree_entry := display_git_tree_entry_t{}
		os_mode, err := entry.Mode.ToOSFileMode()
		if err != nil {
			display_git_tree_entry.Mode = "----"
		} else {
			display_git_tree_entry.Mode = os_mode.String()[:4]
		}
		display_git_tree_entry.Is_file = entry.Mode.IsFile()
		display_git_tree_entry.Size, err = target.Size(entry.Name)
		if err != nil {
			display_git_tree_entry.Size = 0
		}
		display_git_tree_entry.Name = entry.Name
		display_git_tree = append(display_git_tree, display_git_tree_entry)
	}
	data["files"] = display_git_tree

	err = templates.ExecuteTemplate(w, "repo_tree_dir", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
