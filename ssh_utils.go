package main

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var err_ssh_illegal_endpoint = errors.New("Illegal endpoint during SSH access")

func get_repo_path_perms_from_ssh_path_pubkey(ctx context.Context, ssh_path string, ssh_pubkey string) (repo_path string, access bool, err error) {
	segments := strings.Split(strings.TrimPrefix(ssh_path, "/"), "/")

	for i, segment := range segments {
		var err error
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return "", false, err
		}
	}

	if segments[0] == ":" {
		return "", false, err_ssh_illegal_endpoint
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
		return "", false, err_ssh_illegal_endpoint
	case len(segments) <= separator_index+2:
		return "", false, err_ssh_illegal_endpoint
	}

	group_name := segments[0]
	module_type := segments[separator_index+1]
	module_name := segments[separator_index+2]
	switch module_type {
	case "repos":
		return get_path_perm_by_group_repo_key(ctx, group_name, module_name, ssh_pubkey)
	default:
		return "", false, err_ssh_illegal_endpoint
	}
}
