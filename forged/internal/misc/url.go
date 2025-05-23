// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package misc

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrDupRefSpec = errors.New("duplicate ref spec")
	ErrNoRefSpec  = errors.New("no ref spec")
)

// getParamRefTypeName looks at the query parameters in an HTTP request and
// returns its ref name and type, if any.
func GetParamRefTypeName(request *http.Request) (retRefType, retRefName string, err error) {
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
				err = ErrDupRefSpec
				return
			}
			done = true
			if len(refName) != 1 {
				err = ErrDupRefSpec
				return
			}
			retRefName = refName[0]
			retRefType = refType
		}
	}
	if !done {
		err = ErrNoRefSpec
	}
	return
}

// ParseReqURI parses an HTTP request URL, and returns a slice of path segments
// and the query parameters. It handles %2F correctly.
func ParseReqURI(requestURI string) (segments []string, params url.Values, err error) {
	path, paramsStr, _ := strings.Cut(requestURI, "?")

	segments, err = PathToSegments(path)
	if err != nil {
		return
	}

	params, err = url.ParseQuery(paramsStr)
	return
}

// PathToSegments splits a path into unescaped segments. It handles %2F correctly.
func PathToSegments(path string) (segments []string, err error) {
	segments = strings.Split(strings.TrimPrefix(path, "/"), "/")

	for i, segment := range segments {
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return
		}
	}

	return
}

// RedirectDir returns true and redirects the user to a version of the URL with
// a trailing slash, if and only if the request URL does not already have a
// trailing slash.
func RedirectDir(writer http.ResponseWriter, request *http.Request) bool {
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

// RedirectNoDir returns true and redirects the user to a version of the URL
// without a trailing slash, if and only if the request URL has a trailing
// slash.
func RedirectNoDir(writer http.ResponseWriter, request *http.Request) bool {
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

// RedirectUnconditionally unconditionally redirects the user back to the
// current page while preserving query parameters.
func RedirectUnconditionally(writer http.ResponseWriter, request *http.Request) {
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

// SegmentsToURL joins URL segments to the path component of a URL.
// Each segment is escaped properly first.
func SegmentsToURL(segments []string) string {
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

// AnyContain returns true if and only if ss contains a string that contains c.
func AnyContain(ss []string, c string) bool {
	for _, s := range ss {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}
