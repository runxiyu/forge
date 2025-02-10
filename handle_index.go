package main

import (
	"net/http"
	"os"
)

func handle_index(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)

	entries, err := os.ReadDir(config.Git.Root)
	if err != nil {
		w.Write([]byte("Error listing categories: " + err.Error()))
		return
	}

	categories := []string{}
	for _, entry := range entries {
		categories = append(categories, entry.Name())
	}
	data["categories"] = categories

	err = templates.ExecuteTemplate(w, "index", data)
	if err != nil {
		w.Write([]byte("Error rendering template: " + err.Error()))
		return
	}
}
