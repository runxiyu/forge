package main

import (
	"net/http"
)

func handle_group_repos(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name := params["group_name"]

	rows, err := database.Query(r.Context(), "SELECT r.name, COALESCE(r.description, '') FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1;", group_name)
	if err != nil {
		http.Error(w, "Error getting groups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	repos := []struct {
		Name        string
		Description string
	}{}
	for rows.Next() {
		var repoName, repoDescription string
		if err := rows.Scan(&repoName, &repoDescription); err != nil {
			http.Error(w, "Error scanning repo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		repos = append(repos, struct {
			Name        string
			Description string
		}{repoName, repoDescription})
	}
	params["repos"] = repos

	err = templates.ExecuteTemplate(w, "group_repos", params)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
