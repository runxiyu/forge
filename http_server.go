// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

type http_router_t struct{}

func (router *http_router_t) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clog.Info("Incoming HTTP: " + r.RemoteAddr + " " + r.Method + " " + r.RequestURI)

	var segments []string
	var err error
	var non_empty_last_segments_len int
	var params map[string]any
	var separator_index int

	if segments, _, err = parse_request_uri(r.RequestURI); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	non_empty_last_segments_len = len(segments)
	if segments[len(segments)-1] == "" {
		non_empty_last_segments_len--
	}

	if segments[0] == ":" {
		if len(segments) < 2 {
			http.Error(w, "Blank system endpoint", http.StatusNotFound)
			return
		} else if len(segments) == 2 && redirect_with_slash(w, r) {
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

	params["url_segments"] = segments
	params["global"] = global_data
	var _user_id int // 0 for none
	_user_id, params["username"], err = get_user_info_from_request(r)
	if errors.Is(err, http.ErrNoCookie) {
	} else if errors.Is(err, pgx.ErrNoRows) {
	} else if err != nil {
		http.Error(w, "Error getting user info from request: "+err.Error(), http.StatusInternalServerError)
		return
	}

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

	separator_index = -1
	for i, part := range segments {
		if part == ":" {
			separator_index = i
			break
		}
	}

	params["separator_index"] = separator_index

	// TODO
	if separator_index > 1 {
		http.Error(w, "Subgroups haven't been implemented yet", http.StatusNotImplemented)
		return
	}

	var module_type string
	var module_name string
	var group_name string

	switch {
	case non_empty_last_segments_len == 0:
		handle_index(w, r, params)
	case separator_index == -1:
		http.Error(w, "Group indexing hasn't been implemented yet", http.StatusNotImplemented)
	case non_empty_last_segments_len == separator_index+1:
		http.Error(w, "Group root hasn't been implemented yet", http.StatusNotImplemented)
	case non_empty_last_segments_len == separator_index+2:
		if redirect_with_slash(w, r) {
			return
		}
		module_type = segments[separator_index+1]
		params["group_name"] = segments[0]
		switch module_type {
		case "repos":
			handle_group_repos(w, r, params)
		default:
			http.Error(w, fmt.Sprintf("Unknown module type: %s", module_type), http.StatusNotFound)
		}
	default:
		module_type = segments[separator_index+1]
		module_name = segments[separator_index+2]
		group_name = segments[0]
		params["group_name"] = group_name
		switch module_type {
		case "repos":
			params["repo_name"] = module_name

			if non_empty_last_segments_len > separator_index+3 {
				switch segments[separator_index+3] {
				case "info":
					if err = handle_repo_info(w, r, params); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				case "git-upload-pack":
					if err = handle_upload_pack(w, r, params); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}

			if params["ref_type"], params["ref_name"], err = get_param_ref_and_type(r); err != nil {
				if errors.Is(err, err_no_ref_spec) {
					params["ref_type"] = ""
				} else {
					http.Error(w, "Error querying ref type: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// TODO: subgroups

			if params["repo"], params["repo_description"], params["repo_id"], err = open_git_repo(r.Context(), group_name, module_name); err != nil {
				http.Error(w, "Error opening repo: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if non_empty_last_segments_len == separator_index+3 {
				if redirect_with_slash(w, r) {
					return
				}
				handle_repo_index(w, r, params)
				return
			}

			repo_feature := segments[separator_index+3]
			switch repo_feature {
			case "tree":
				params["rest"] = strings.Join(segments[separator_index+4:], "/")
				if len(segments) < separator_index+5 && redirect_with_slash(w, r) {
					return
				}
				handle_repo_tree(w, r, params)
			case "raw":
				params["rest"] = strings.Join(segments[separator_index+4:], "/")
				if len(segments) < separator_index+5 && redirect_with_slash(w, r) {
					return
				}
				handle_repo_raw(w, r, params)
			case "log":
				if non_empty_last_segments_len > separator_index+4 {
					http.Error(w, "Too many parameters", http.StatusBadRequest)
					return
				}
				if redirect_with_slash(w, r) {
					return
				}
				handle_repo_log(w, r, params)
			case "commit":
				if redirect_without_slash(w, r) {
					return
				}
				params["commit_id"] = segments[separator_index+4]
				handle_repo_commit(w, r, params)
			case "contrib":
				if redirect_with_slash(w, r) {
					return
				}
				switch non_empty_last_segments_len {
				case separator_index + 4:
					handle_repo_contrib_index(w, r, params)
				case separator_index + 5:
					params["mr_id"] = segments[separator_index+4]
					handle_repo_contrib_one(w, r, params)
				default:
					http.Error(w, "Too many parameters", http.StatusBadRequest)
				}
			default:
				http.Error(w, fmt.Sprintf("Unknown repo feature: %s", repo_feature), http.StatusNotFound)
			}
		default:
			http.Error(w, fmt.Sprintf("Unknown module type: %s", module_type), http.StatusNotFound)
		}
	}
}
