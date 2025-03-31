// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var (
	errDupRefSpec = errors.New("duplicate ref spec")
	errNoRefSpec  = errors.New("no ref spec")
)

func getParamRefTypeName(request *http.Request) (retRefType, retRefName string, err error) {
	rawQuery := request.URL.RawQuery
	queryValues, err := url.ParseQuery(rawQuery)
	if err != nil {
		return
	}
	done := false
	for _, refType := range []string{"commit", "branch", "tag"} {
		refName, ok := queryValues[refType]
		if ok {
			if done {
				err = errDupRefSpec
				return
			}
			done = true
			if len(refName) != 1 {
				err = errDupRefSpec
				return
			}
			retRefName = refName[0]
			retRefType = refType
		}
	}
	if !done {
		err = errNoRefSpec
	}
	return
}

func parseReqURI(requestURI string) (segments []string, params url.Values, err error) {
	path, paramsStr, _ := strings.Cut(requestURI, "?")

	segments = strings.Split(strings.TrimPrefix(path, "/"), "/")

	for i, segment := range segments {
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return
		}
	}

	params, err = url.ParseQuery(paramsStr)
	return
}

func redirectDir(writer http.ResponseWriter, request *http.Request) bool {
	requestURI := request.RequestURI

	pathEnd := strings.IndexAny(requestURI, "?#")
	var path, rest string
	if pathEnd == -1 {
		path = requestURI
	} else {
		path = requestURI[:pathEnd]
		rest = requestURI[pathEnd:]
	}

	if !strings.HasSuffix(path, "/") {
		http.Redirect(writer, request, path+"/"+rest, http.StatusSeeOther)
		return true
	}
	return false
}

func redirectNoDir(writer http.ResponseWriter, request *http.Request) bool {
	requestURI := request.RequestURI

	pathEnd := strings.IndexAny(requestURI, "?#")
	var path, rest string
	if pathEnd == -1 {
		path = requestURI
	} else {
		path = requestURI[:pathEnd]
		rest = requestURI[pathEnd:]
	}

	if strings.HasSuffix(path, "/") {
		http.Redirect(writer, request, strings.TrimSuffix(path, "/")+rest, http.StatusSeeOther)
		return true
	}
	return false
}

func redirectUnconditionally(writer http.ResponseWriter, request *http.Request) {
	requestURI := request.RequestURI

	pathEnd := strings.IndexAny(requestURI, "?#")
	var path, rest string
	if pathEnd == -1 {
		path = requestURI
	} else {
		path = requestURI[:pathEnd]
		rest = requestURI[pathEnd:]
	}

	http.Redirect(writer, request, path+rest, http.StatusSeeOther)
}

func segmentsToURL(segments []string) string {
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func anyContain(ss []string, c string) bool {
	for _, s := range ss {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}
