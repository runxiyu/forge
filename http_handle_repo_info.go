package main

import (
	"net/http"
	"net/url"
)

func handle_repo_info(w http.ResponseWriter, r *http.Request, params map[string]any) {
	http.Error(w, "\x1b[1;93mHi! We do not support Git operations over HTTP yet.\x1b[0m\n\x1b[1;93mMeanwhile, please use ssh by simply replacing the scheme with \"ssh://\":\x1b[0m\n\x1b[1;93mssh://"+r.Host+"/"+url.PathEscape(params["group_name"].(string))+"/:/repos/"+url.PathEscape(params["repo_name"].(string))+"\x1b[0m", http.StatusNotImplemented)
}
