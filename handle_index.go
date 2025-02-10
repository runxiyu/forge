package main

import (
	"net/http"
	"os"
)

func handle_index(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)

	entries, err := os.ReadDir(config.Git.Root)
	if err != nil {
		_, _ = w.Write([]byte("Error listing groups: " + err.Error()))
		return
	}

	groups := []string{}
	for _, entry := range entries {
		groups = append(groups, entry.Name())
	}
	data["groups"] = groups

	err = templates.ExecuteTemplate(w, "index", data)
	if err != nil {
		_, _ = w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
