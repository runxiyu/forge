package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"strconv"
)

type http_router_t struct{}

func (router *http_router_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	segments, _, err := parse_request_uri(r.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	non_empty_last_segments_len := len(segments)
	dir_mode := false
	if segments[len(segments)-1] == "" {
		non_empty_last_segments_len--
		dir_mode = true
	}

	if segments[0] == ":" {
		if len(segments) < 2 {
			http.Error(w, "Blank system endpoint", http.StatusNotFound)
			return
		} else if len(segments) == 2 && !dir_mode {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
			return
		}

		switch segments[1] {
		case "static":
			static_handler.ServeHTTP(w, r)
			return
		case "source":
			source_handler.ServeHTTP(w, r)
			return
		}
	}

	params := make(map[string]any)
	params["global"] = global_data
	var _user_id int
	_user_id, params["username"], err = get_user_info_from_request(r)
	if _user_id == 0 {
		params["user_id"] = ""
	} else {
		params["user_id"] = strconv.Itoa(_user_id)
	}

	if segments[0] == ":" {
		switch segments[1] {
		case "login":
			handle_login(w, r, params)
			return
		case "users":
			handle_users(w, r, params)
			return
		default:
			http.Error(w, fmt.Sprintf("Unknown system module type: %s", segments[1]), http.StatusNotFound)
			return
		}
	}

	separator_index := -1
	for i, part := range segments {
		if part == ":" {
			separator_index = i
			break
		}
	}

	switch {
	case non_empty_last_segments_len == 0:
		handle_index(w, r, params)
	case separator_index == -1:
		http.Error(w, "Group indexing hasn't been implemented yet", http.StatusNotImplemented)
	case non_empty_last_segments_len == separator_index+1:
		http.Error(w, "Group root hasn't been implemented yet", http.StatusNotImplemented)
	case non_empty_last_segments_len == separator_index+2:
		if !dir_mode {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
			return
		}
		module_type := segments[separator_index+1]
		params["group_name"] = segments[0]
		switch module_type {
		case "repos":
			handle_group_repos(w, r, params)
		default:
			http.Error(w, fmt.Sprintf("Unknown module type: %s", module_type), http.StatusNotFound)
		}
	default:
		module_type := segments[separator_index+1]
		module_name := segments[separator_index+2]
		params["group_name"] = segments[0]
		switch module_type {
		case "repos":
			params["repo_name"] = module_name
			// TODO: subgroups
			if non_empty_last_segments_len == separator_index+3 {
				if !dir_mode {
					http.Redirect(w, r, r.URL.Path+"/", http.StatusSeeOther)
					return
				}
				handle_repo_index(w, r, params)
				return
			}
			repo_feature := segments[separator_index+3]
			switch repo_feature {
			case "info":
				handle_repo_info(w, r, params)
			case "tree":
				params["rest"] = strings.Join(segments[separator_index+4:], "/")
				handle_repo_tree(w, r, params)
			case "raw":
				params["rest"] = strings.Join(segments[separator_index+4:], "/")
				handle_repo_raw(w, r, params)
			case "log":
				if non_empty_last_segments_len != separator_index+5 {
					http.Error(w, "Too many parameters", http.StatusBadRequest)
					return
				}
				if dir_mode {
					http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusSeeOther)
					return
				}
				params["ref"] = segments[separator_index+4]
				handle_repo_log(w, r, params)
			case "commit":
				if dir_mode {
					http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusSeeOther)
					return
				}
				params["commit_id"] = segments[separator_index+4]
				handle_repo_commit(w, r, params)
			default:
				http.Error(w, fmt.Sprintf("Unknown repo feature: %s", repo_feature), http.StatusNotFound)
			}
		default:
			http.Error(w, fmt.Sprintf("Unknown module type: %s", module_type), http.StatusNotFound)
		}
	}
}

var err_bad_request = errors.New("Bad Request")
