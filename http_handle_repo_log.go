package main

import (
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

// TODO: I probably shouldn't include *all* commits here...
func handle_repo_log(w http.ResponseWriter, r *http.Request, params map[string]any) {
	group_name, repo_name, ref_name := params["group_name"].(string), params["repo_name"].(string), params["ref"].(string)
	repo, err := open_git_repo(r.Context(), group_name, repo_name)
	if err != nil {
		fmt.Fprintln(w, "Error opening repo:", err.Error())
		return
	}
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(ref_name), true)
	if err != nil {
		fmt.Fprintln(w, "Error getting repo reference:", err.Error())
		return
	}
	ref_hash := ref.Hash()
	commits, err := get_recent_commits(repo, ref_hash, -1)
	if err != nil {
		fmt.Fprintln(w, "Error getting recent commits:", err.Error())
		return
	}
	params["commits"] = commits

	err = templates.ExecuteTemplate(w, "repo_log", params)
	if err != nil {
		fmt.Fprintln(w, "Error rendering template:", err.Error())
		return
	}
}
