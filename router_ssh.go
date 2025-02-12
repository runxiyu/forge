package main

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var err_ssh_illegal_endpoint = errors.New("Illegal endpoint during SSH access")

func get_repo_path_from_ssh_path(ctx context.Context, ssh_path string) (repo_path string, err error) {
	segments := strings.Split(strings.TrimPrefix(ssh_path, "/"), "/")

	for i, segment := range segments {
		var err error
		segments[i], err = url.QueryUnescape(segment)
		if err != nil {
			return "", err
		}
	}

	if segments[0] == ":" {
		return "", err_ssh_illegal_endpoint
	}

	separator_index := -1
	for i, part := range segments {
		if part == ":" {
			separator_index = i
			break
		}
	}
	if segments[len(segments)-1] == "" {
		segments = segments[:len(segments)-1]
	}

	switch {
	case separator_index == -1:
		return "", err_ssh_illegal_endpoint
	case len(segments) <= separator_index+2:
		return "", err_ssh_illegal_endpoint
	}

	group_name := segments[0]
	module_type := segments[separator_index+1]
	module_name := segments[separator_index+2]
	switch module_type {
	case "repos":
		var fs_path string
		err := database.QueryRow(ctx, "SELECT r.filesystem_path FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1 AND r.name = $2;", group_name, module_name).Scan(&fs_path)
		return fs_path, err
	default:
		return "", err_ssh_illegal_endpoint
	}
}
