// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"go.lindenii.runxiyu.org/lindenii-common/ansiec"
)

var err_ssh_illegal_endpoint = errors.New("illegal endpoint during SSH access")

func get_repo_path_perms_from_ssh_path_pubkey(ctx context.Context, ssh_path string, ssh_pubkey string) (group_path []string, repo_name string, repo_id int, repo_path string, direct_access bool, contrib_requirements string, user_type string, user_id int, err error) {
	var segments []string
	var separator_index int
	var module_type, module_name string

	segments = strings.Split(strings.TrimPrefix(ssh_path, "/"), "/")

	for i, segment := range segments {
		var err error
		segments[i], err = url.PathUnescape(segment)
		if err != nil {
			return []string{}, "", 0, "", false, "", "", 0, err
		}
	}

	if segments[0] == ":" {
		return []string{}, "", 0, "", false, "", "", 0, err_ssh_illegal_endpoint
	}

	separator_index = -1
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
		return []string{}, "", 0, "", false, "", "", 0, err_ssh_illegal_endpoint
	case len(segments) <= separator_index+2:
		return []string{}, "", 0, "", false, "", "", 0, err_ssh_illegal_endpoint
	}

	group_path = segments[:separator_index]
	module_type = segments[separator_index+1]
	module_name = segments[separator_index+2]
	repo_name = module_name
	switch module_type {
	case "repos":
		_1, _2, _3, _4, _5, _6, _7 := get_path_perm_by_group_repo_key(ctx, group_path, module_name, ssh_pubkey)
		return group_path, repo_name, _1, _2, _3, _4, _5, _6, _7
	default:
		return []string{}, "", 0, "", false, "", "", 0, err_ssh_illegal_endpoint
	}
}

func wf_error(w io.Writer, format string, args ...any) {
	fmt.Fprintln(w, ansiec.Red+fmt.Sprintf(format, args...)+ansiec.Reset)
}
