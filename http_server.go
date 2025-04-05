// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/forge/misc"
)

// ServeHTTP handles all incoming HTTP requests and routes them to the correct
// location.
//
// TODO: This function is way too large.
func (s *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var remoteAddr string
	if s.config.HTTP.ReverseProxy {
		remoteAddrs, ok := request.Header["X-Forwarded-For"]
		if ok && len(remoteAddrs) == 1 {
			remoteAddr = remoteAddrs[0]
		} else {
			remoteAddr = request.RemoteAddr
		}
	} else {
		remoteAddr = request.RemoteAddr
	}
	slog.Info("incoming http", "addr", remoteAddr, "method", request.Method, "uri", request.RequestURI)

	var segments []string
	var err error
	var sepIndex int
	params := make(map[string]any)

	if segments, _, err = misc.ParseReqURI(request.RequestURI); err != nil {
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
	params["global"] = s.globalData
	var userID int // 0 for none
	userID, params["username"], err = s.getUserFromRequest(request)
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

	for _, v := range segments {
		if strings.Contains(v, ":") {
			errorPage400Colon(writer, params)
			return
		}
	}

	if len(segments) == 0 {
		s.httpHandleIndex(writer, request, params)
		return
	}

	if segments[0] == "-" {
		if len(segments) < 2 {
			errorPage404(writer, params)
			return
		} else if len(segments) == 2 && misc.RedirectDir(writer, request) {
			return
		}

		switch segments[1] {
		case "static":
			s.staticHandler.ServeHTTP(writer, request)
			return
		case "source":
			s.sourceHandler.ServeHTTP(writer, request)
			return
		}
	}

	if segments[0] == "-" {
		switch segments[1] {
		case "login":
			s.httpHandleLogin(writer, request, params)
			return
		case "users":
			httpHandleUsers(writer, request, params)
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
		if misc.RedirectDir(writer, request) {
			return
		}
		s.httpHandleGroupIndex(writer, request, params)
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
					if err = s.httpHandleRepoInfo(writer, request, params); err != nil {
						errorPage500(writer, params, err.Error())
					}
					return
				case "git-upload-pack":
					if err = s.httpHandleUploadPack(writer, request, params); err != nil {
						errorPage500(writer, params, err.Error())
					}
					return
				}
			}

			if params["ref_type"], params["ref_name"], err = misc.GetParamRefTypeName(request); err != nil {
				if errors.Is(err, misc.ErrNoRefSpec) {
					params["ref_type"] = ""
				} else {
					errorPage400(writer, params, "Error querying ref type: "+err.Error())
					return
				}
			}

			if params["repo"], params["repo_description"], params["repo_id"], _, err = s.openRepo(request.Context(), groupPath, moduleName); err != nil {
				errorPage500(writer, params, "Error opening repo: "+err.Error())
				return
			}

			repoURLRoot := "/"
			for _, part := range segments[:sepIndex+3] {
				repoURLRoot = repoURLRoot + url.PathEscape(part) + "/"
			}
			params["repo_url_root"] = repoURLRoot
			params["repo_patch_mailing_list"] = repoURLRoot[1:len(repoURLRoot)-1] + "@" + s.config.LMTP.Domain
			params["http_clone_url"] = s.genHTTPRemoteURL(groupPath, moduleName)
			params["ssh_clone_url"] = s.genSSHRemoteURL(groupPath, moduleName)

			if len(segments) == sepIndex+3 {
				if misc.RedirectDir(writer, request) {
					return
				}
				s.httpHandleRepoIndex(writer, request, params)
				return
			}

			repoFeature := segments[sepIndex+3]
			switch repoFeature {
			case "tree":
				if misc.AnyContain(segments[sepIndex+4:], "/") {
					errorPage400(writer, params, "Repo tree paths may not contain slashes in any segments")
					return
				}
				if dirMode {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/") + "/"
				} else {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/")
				}
				if len(segments) < sepIndex+5 && misc.RedirectDir(writer, request) {
					return
				}
				s.httpHandleRepoTree(writer, request, params)
			case "branches":
				if misc.RedirectDir(writer, request) {
					return
				}
				s.httpHandleRepoBranches(writer, request, params)
				return
			case "raw":
				if misc.AnyContain(segments[sepIndex+4:], "/") {
					errorPage400(writer, params, "Repo tree paths may not contain slashes in any segments")
					return
				}
				if dirMode {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/") + "/"
				} else {
					params["rest"] = strings.Join(segments[sepIndex+4:], "/")
				}
				if len(segments) < sepIndex+5 && misc.RedirectDir(writer, request) {
					return
				}
				s.httpHandleRepoRaw(writer, request, params)
			case "log":
				if len(segments) > sepIndex+4 {
					errorPage400(writer, params, "Too many parameters")
					return
				}
				if misc.RedirectDir(writer, request) {
					return
				}
				httpHandleRepoLog(writer, request, params)
			case "commit":
				if len(segments) != sepIndex+5 {
					errorPage400(writer, params, "Incorrect number of parameters")
					return
				}
				if misc.RedirectNoDir(writer, request) {
					return
				}
				params["commit_id"] = segments[sepIndex+4]
				httpHandleRepoCommit(writer, request, params)
			case "contrib":
				if misc.RedirectDir(writer, request) {
					return
				}
				switch len(segments) {
				case sepIndex + 4:
					s.httpHandleRepoContribIndex(writer, request, params)
				case sepIndex + 5:
					params["mr_id"] = segments[sepIndex+4]
					s.httpHandleRepoContribOne(writer, request, params)
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
