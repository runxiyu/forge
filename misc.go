package main

import (
	"errors"
	"strings"
)

type name_desc_t struct {
	Name        string
	Description string
}

var err_environ_no_separator = errors.New("No separator found in environ line")

func environ_to_map(environ_strings []string) (result map[string]string, err error) {
	for _, environ_string := range environ_strings {
		key, value, found := strings.Cut(environ_string, "=")
		if !found {
			return result, err_environ_no_separator
		}
		result[key] = value
	}
	return result, err_environ_no_separator
}
