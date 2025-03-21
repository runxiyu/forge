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

type forgeHTTPRouter struct{}

func (router *forgeHTTPRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clog.Info("Incoming HTTP: " + r.RemoteAddr + " " + r.Method + " " + r.RequestURI)

	var segments []string
	var err error
	var contentfulSegmentsLen int
	var sepIndex int
	params := make(map[string]any)

	if segments, _, err = parseReqURI(r.RequestURI); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	contentfulSegmentsLen = len(segments)
	if segments[len(segments)-1] == "" {
		contentfulSegmentsLen--
	}

	if segments[0] == ":" {
		if len(segments) < 2 {
			http.Error(w, "Blank system endpoint", http.StatusNotFound)
			return
		} else if len(segments) == 2 && redirectDir(w, r) {
			return
		}

		switch segments[1] {
		case "static":
			staticHandler.ServeHTTP(w, r)
			return
		case "source":
			sourceHandler.ServeHTTP(w, r)
			return
		}
	}

	params["url_segments"] = segments
	params["global"] = globalData
	var userID int // 0 for none
	userID, params["username"], err = getUserFromRequest(r)
	params["user_id"] = userID
	if errors.Is(err, http.ErrNoCookie) {
	} else if errors.Is(err, pgx.ErrNoRows) {
	} else if err != nil {
		http.Error(w, "Error getting user info from request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if userID == 0 {
		params["user_id_string"] = ""
	} else {
		params["user_id_string"] = strconv.Itoa(userID)
	}

	if segments[0] == ":" {
		switch segments[1] {
		case "login":
			httpHandleLogin(w, r, params)
			return
		case "users":
			httpHandleUsers(w, r, params)
			return
		case "gc":
			httpHandleGC(w, r, params)
			return
		default:
			http.Error(w, fmt.Sprintf("Unknown system module type: %s", segments[1]), http.StatusNotFound)
			return
		}
	}

	sepIndex = -1
	for i, part := range segments {
		if part == ":" {
			sepIndex = i
			break
		}
	}

	params["separator_index"] = sepIndex

	var groupPath []string
	var moduleType string
	var moduleName string

	if sepIndex > 0 {
		groupPath = segments[:sepIndex]
	} else {
		groupPath = segments[:len(segments)-1]
	}
	params["group_path"] = groupPath

	switch {
	case contentfulSegmentsLen == 0:
		httpHandleIndex(w, r, params)
	case sepIndex == -1:
		if redirectDir(w, r) {
			return
		}
		httpHandleGroupIndex(w, r, params)
	case contentfulSegmentsLen == sepIndex+1:
		http.Error(w, "Illegal path 1", http.StatusNotImplemented)
		return
	case contentfulSegmentsLen == sepIndex+2:
		http.Error(w, "Illegal path 2", http.StatusNotImplemented)
		return
	default:
		moduleType = segments[sepIndex+1]
		moduleName = segments[sepIndex+2]
		switch moduleType {
		case "repos":
			params["repo_name"] = moduleName

			if contentfulSegmentsLen > sepIndex+3 {
				switch segments[sepIndex+3] {
				case "info":
					if err = httpHandleRepoInfo(w, r, params); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				case "git-upload-pack":
					if err = httpHandleUploadPack(w, r, params); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}

			if params["ref_type"], params["ref_name"], err = getParamRefTypeName(r); err != nil {
				if errors.Is(err, errNoRefSpec) {
					params["ref_type"] = ""
				} else {
					http.Error(w, "Error querying ref type: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// TODO: subgroups

			if params["repo"], params["repo_description"], params["repo_id"], err = openRepo(r.Context(), groupPath, moduleName); err != nil {
				http.Error(w, "Error opening repo: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if contentfulSegmentsLen == sepIndex+3 {
				if redirectDir(w, r) {
					return
				}
				httpHandleRepoIndex(w, r, params)
				return
			}

			repoFeature := segments[sepIndex+3]
			switch repoFeature {
			case "tree":
				params["rest"] = strings.Join(segments[sepIndex+4:], "/")
				if len(segments) < sepIndex+5 && redirectDir(w, r) {
					return
				}
				httpHandleRepoTree(w, r, params)
			case "raw":
				params["rest"] = strings.Join(segments[sepIndex+4:], "/")
				if len(segments) < sepIndex+5 && redirectDir(w, r) {
					return
				}
				httpHandleRepoRaw(w, r, params)
			case "log":
				if contentfulSegmentsLen > sepIndex+4 {
					http.Error(w, "Too many parameters", http.StatusBadRequest)
					return
				}
				if redirectDir(w, r) {
					return
				}
				httpHandleRepoLog(w, r, params)
			case "commit":
				if redirectNoDir(w, r) {
					return
				}
				params["commit_id"] = segments[sepIndex+4]
				httpHandleRepoCommit(w, r, params)
			case "contrib":
				if redirectDir(w, r) {
					return
				}
				switch contentfulSegmentsLen {
				case sepIndex + 4:
					httpHandleRepoContribIndex(w, r, params)
				case sepIndex + 5:
					params["mr_id"] = segments[sepIndex+4]
					httpHandleRepoContribOne(w, r, params)
				default:
					http.Error(w, "Too many parameters", http.StatusBadRequest)
				}
			default:
				http.Error(w, fmt.Sprintf("Unknown repo feature: %s", repoFeature), http.StatusNotFound)
			}
		default:
			http.Error(w, fmt.Sprintf("Unknown module type: %s", moduleType), http.StatusNotFound)
		}
	}
}
