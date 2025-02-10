package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handle_category_repos(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	category_name := r.PathValue("category_name")
	data["category_name"] = category_name
	entries, err := os.ReadDir(filepath.Join(config.Git.Root, category_name))
	if err != nil {
		_, _ = w.Write([]byte("Error listing repos: " + err.Error()))
		return
	}

	repos := []string{}
	for _, entry := range entries {
		this_name := entry.Name()
		if strings.HasSuffix(this_name, ".git") {
			repos = append(repos, strings.TrimSuffix(this_name, ".git"))
		}
	}
	data["repos"] = repos

	err = templates.ExecuteTemplate(w, "category_repos", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
