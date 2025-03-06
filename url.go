// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var (
	err_duplicate_ref_spec = errors.New("duplicate ref spec")
	err_no_ref_spec        = errors.New("no ref spec")
)

func get_param_ref_and_type(r *http.Request) (ref_type, ref string, err error) {
	qr := r.URL.RawQuery
	q, err := url.ParseQuery(qr)
	if err != nil {
		return
	}
	done := false
	for _, _ref_type := range []string{"commit", "branch", "tag"} {
		_ref, ok := q[_ref_type]
		if ok {
			if done {
				err = err_duplicate_ref_spec
				return
			} else {
				done = true
				if len(_ref) != 1 {
					err = err_duplicate_ref_spec
					return
				}
				ref = _ref[0]
				ref_type = _ref_type
			}
		}
	}
	if !done {
		err = err_no_ref_spec
	}
	return
}

func parse_request_uri(request_uri string) (segments []string, params url.Values, err error) {
	path, params_string, _ := strings.Cut(request_uri, "?")

	segments = strings.Split(strings.TrimPrefix(path, "/"), "/")

	for i, segment := range segments {
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return
		}
	}

	params, err = url.ParseQuery(params_string)
	return
}

func redirect_with_slash(w http.ResponseWriter, r *http.Request) bool {
	request_uri := r.RequestURI

	path_end := strings.IndexAny(request_uri, "?#")
	var path, rest string
	if path_end == -1 {
		path = request_uri
	} else {
		path = request_uri[:path_end]
		rest = request_uri[path_end:]
	}

	if !strings.HasSuffix(path, "/") {
		http.Redirect(w, r, path+"/"+rest, http.StatusSeeOther)
		return true
	}
	return false
}

func redirect_without_slash(w http.ResponseWriter, r *http.Request) bool {
	request_uri := r.RequestURI

	path_end := strings.IndexAny(request_uri, "?#")
	var path, rest string
	if path_end == -1 {
		path = request_uri
	} else {
		path = request_uri[:path_end]
		rest = request_uri[path_end:]
	}

	if strings.HasSuffix(path, "/") {
		http.Redirect(w, r, strings.TrimSuffix(path, "/")+rest, http.StatusSeeOther)
		return true
	}
	return false
}

func path_escape_cat_segments(segments []string) string {
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}
