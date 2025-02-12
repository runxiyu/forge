package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"go.lindenii.runxiyu.org/lindenii-common/misc"
)

var (
	err_duplicate_ref_spec = errors.New("Duplicate ref spec")
	err_no_ref_spec        = errors.New("No ref spec")
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
		segments[i], _ = url.QueryUnescape(segment)
	}

	params, err = url.ParseQuery(params_string)
	if err != nil {
		return nil, nil, misc.Wrap_one_error(err_bad_request, err)
	}

	return
}
