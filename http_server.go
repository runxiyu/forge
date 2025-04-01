// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
)

type forgeHTTPRouter struct{}

// ServeHTTP handles all incoming HTTP requests and routes them to the correct
// location.
//
// TODO: This function is way too large.
func (router *forgeHTTPRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var remoteAddr string
	if config.HTTP.ReverseProxy {
		remoteAddrs, ok := request.Header["X-Forwarded-For"]
		if ok && len(remoteAddrs) == 1 {
			remoteAddr = remoteAddrs[0]
		} else {
			remoteAddr = request.RemoteAddr
		}
	} else {
		remoteAddr = request.RemoteAddr
	}
	clog.Info("Incoming HTTP: " + remoteAddr + " " + request.Method + " " + request.RequestURI)

	var segments []string
	var err error
	var sepIndex int
	params := make(map[string]any)

	if segments, _, err = parseReqURI(request.RequestURI); err != nil {
		errorPage400(writer, params, "Error parsing request URI: "+err.Error())
		return
	}
	dirMode := false
	if segments[len(segments)-1] == "" {
		dirMode = true
		segments = segments[:len(segments)-1]
	}

	params["url_segments"] = segments
	params["dir_mode"] = dirMode
	params["global"] = globalData
	var userID int // 0 for none
	userID, params["username"], err = getUserFromRequest(request)
	params["user_id"] = userID
	if err != nil && !errors.Is(err, http.ErrNoCookie) && !errors.Is(err, pgx.ErrNoRows) {
		errorPage500(writer, params, "Error getting user info from request: "+err.Error())
		return
	}

	if userID == 0 {
		params["user_id_string"] = ""
	} else {
		params["user_id_string"] = strconv.Itoa(userID)
	}

	if len(segments) == 0 {
		httpHandleIndex(writer, request, params)
		return
	}

	if segments[0] == "-" {
		if len(segments) < 2 {
			errorPage404(writer, params)
			return
		} else if len(segments) == 2 && redirectDir(writer, request) {
			return
		}

		switch segments[1] {
		case "man":
			manHandler.ServeHTTP(writer, request)
			return
		case "static":
			staticHandler.ServeHTTP(writer, request)
			return
		case "source":
			sourceHandler.ServeHTTP(writer, request)
			return
		}
	}

	if segments[0] == "-" {
		switch segments[1] {
		case "login":
			httpHandleLogin(writer, request, params)
			return
		case "users":
			httpHandleUsers(writer, request, params)
			return
		case "gc":
			httpHandleGC(writer, request, params)
			return
		default:
			errorPage404(writer, params)
			return
		}
	}

	sepIndex = -1
	for i, part := range segments {
		if part == "-" {
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
		groupPath = segments
	}
	params["group_path"] = groupPath

	switch {
	case sepIndex == -1:
		if redirectDir(writer, request) {
			return
		}
		httpHandleGroupIndex(writer, request, params)
	case len(segments) == sepIndex+1:
		errorPage404(writer, params)
		return
	case len(segments) == sepIndex+2:
		errorPage404(writer, params)
		return
	default:
		moduleType = segments[sepIndex+1]
		moduleName = segments[sepIndex+2]
		switch moduleType {
		case "repos":
			params["repo_name"] = moduleName

			if len(segments) > sepIndex+3 {
				switch segments[sepIndex+3] {
				case "info":
					if err = httpHandleRepoInfo(writer, request, params); err != nil {
						errorPage500(writer, params, err.Error())
					}
					return
				case "git-upload-pack":
					if err = httpHandleUploadPack(writer, request, params); err != nil {
						errorPage500(writer, params, err.Error())
					}
					return
				}
			}

			if params["ref_type"], params["ref_name"], err = getParamRefTypeName(request); err != nil {
				if errors.Is(err, errNoRefSpec) {
					params["ref_type"] = ""
				} else {
					errorPage500(writer, params, "Error querying ref type: "+err.Error())
					return
				}
			}

			// TODO: subgroups

			if params["repo"], params["repo_description"], params["repo_id"], err = openRepo(request.Context(), groupPath, moduleName); err != nil {
				errorPage500(writer, params, "Error opening repo: "+err.Error())
				return
			}

			if len(segments) == sepIndex+3 {
				if redirectDir(writer, request) {
					return
				}
				httpHandleRepoIndex(writer, request, params)
				return
			}

			repoFeature := segments[sepIndex+3]
			switch repoFeature {
			case "tree":
				if anyContain(segments[sepIndex+4:], "/") {
					errorPage400(writer, params, "Repo tree paths may not contain slashes in any segments")
					return
				}
				if dirMode {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/") + "/"
				} else {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/")
				}
				if len(segments) < sepIndex+5 && redirectDir(writer, request) {
					return
				}
				httpHandleRepoTree(writer, request, params)
			case "branches":
				if redirectDir(writer, request) {
					return
				}
				httpHandleRepoBranches(writer, request, params)
				return
			case "raw":
				if anyContain(segments[sepIndex+4:], "/") {
					errorPage400(writer, params, "Repo tree paths may not contain slashes in any segments")
					return
				}
				if dirMode {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/") + "/"
				} else {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/")
				}
				if len(segments) < sepIndex+5 && redirectDir(writer, request) {
					return
				}
				httpHandleRepoRaw(writer, request, params)
			case "log":
				if len(segments) > sepIndex+4 {
					errorPage400(writer, params, "Too many parameters")
					return
				}
				if redirectDir(writer, request) {
					return
				}
				httpHandleRepoLog(writer, request, params)
			case "commit":
				if len(segments) != sepIndex+5 {
					errorPage400(writer, params, "Incorrect number of parameters")
					return
				}
				if redirectNoDir(writer, request) {
					return
				}
				params["commit_id"] = segments[sepIndex+4]
				httpHandleRepoCommit(writer, request, params)
			case "contrib":
				if redirectDir(writer, request) {
					return
				}
				switch len(segments) {
				case sepIndex + 4:
					httpHandleRepoContribIndex(writer, request, params)
				case sepIndex + 5:
					params["mr_id"] = segments[sepIndex+4]
					httpHandleRepoContribOne(writer, request, params)
				default:
					errorPage400(writer, params, "Too many parameters")
				}
			default:
				errorPage404(writer, params)
				return
			}
		default:
			errorPage404(writer, params)
			return
		}
	}
}
