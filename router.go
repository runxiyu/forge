package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type http_router_t struct{}

func (router *http_router_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	segments, _, err := parse_request_uri(r.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if segments[0] == ":" {
		switch segments[1] {
		case "static":
			static_handler.ServeHTTP(w, r)
		case "source":
			source_handler.ServeHTTP(w, r)
		default:
			fmt.Fprintln(w, "Unknown system module type:", segments[1])
		}
		return
	}

	separator_index := -1
	for i, part := range segments {
		if part == ":" {
			separator_index = i
			break
		}
	}
	non_empty_last_segments_len := len(segments)
	dir_mode := false
	if segments[len(segments)-1] == "" {
		non_empty_last_segments_len--
		dir_mode = true
	}

	params := make(map[string]string)
	_ = params
	switch {
	case non_empty_last_segments_len == 0:
		handle_index(w, r)
	case separator_index == -1:
		fmt.Fprintln(w, "Group indexing hasn't been implemented yet")
	case non_empty_last_segments_len == separator_index+1:
		fmt.Fprintln(w, "Group root hasn't been implemented yet")
	case non_empty_last_segments_len == separator_index+2:
		module_type := segments[separator_index+1]
		params["group_name"] = segments[0]
		switch module_type {
		case "repos":
			handle_group_repos(w, r, params)
		default:
			fmt.Fprintln(w, "Unknown module type:", module_type)
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
			case "tree":
				params["rest"] = strings.Join(segments[separator_index+4:], "/")
				handle_repo_tree(w, r, params)
			case "raw":
				params["rest"] = strings.Join(segments[separator_index+4:], "/")
				handle_repo_raw(w, r, params)
			case "log":
				params["ref"] = segments[separator_index+4]
				handle_repo_log(w, r, params)
			case "commit":
				params["commit_id"] = segments[separator_index+4]
				handle_repo_commit(w, r, params)
			}
		default:
			fmt.Fprintln(w, "Unknown module type:", module_type)
		}
	}
}

var err_bad_request = errors.New("Bad Request")

