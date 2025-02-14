package main

import (
	"net/http"
)

func handle_repo_info(w http.ResponseWriter, r *http.Request, params map[string]any) {
	http.Error(w, "\x1b[1;93mHi! We do not support Git operations over HTTP yet.\x1b[0m\n\x1b[1;93mMeanwhile, please use SSH (setupless anonymous access enabled):\x1b[0m\n\x1b[1;93m"+generate_ssh_remote_url(params["group_name"].(string), params["repo_name"].(string))+"\x1b[0m", http.StatusNotImplemented)
}
