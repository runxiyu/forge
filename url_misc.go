package main

import (
	"errors"
	"net/http"
	"net/url"
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
