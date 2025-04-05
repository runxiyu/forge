// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"strings"

	"git.sr.ht/~sircmpwn/go-bare"
)

// httpHandleRepoRaw serves raw files, or directory listings that point to raw
// files.
func httpHandleRepoRaw(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	repoName := params["repo_name"].(string)
	groupPath := params["group_path"].([]string)
	rawPathSpec := params["rest"].(string)
	pathSpec := strings.TrimSuffix(rawPathSpec, "/")
	params["path_spec"] = pathSpec

	_, repoPath, _, _, _, _, _ := getRepoInfo(request.Context(), groupPath, repoName, "")

	conn, err := net.Dial("unix", config.Git.Socket)
	if err != nil {
		errorPage500(writer, params, "git2d connection failed: "+err.Error())
		return
	}
	defer conn.Close()

	brWriter := bare.NewWriter(conn)
	brReader := bare.NewReader(conn)

	if err := brWriter.WriteData([]byte(repoPath)); err != nil {
		errorPage500(writer, params, "sending repo path failed: "+err.Error())
		return
	}
	if err := brWriter.WriteUint(2); err != nil {
		errorPage500(writer, params, "sending command failed: "+err.Error())
		return
	}
	if err := brWriter.WriteData([]byte(pathSpec)); err != nil {
		errorPage500(writer, params, "sending path failed: "+err.Error())
		return
	}

	status, err := brReader.ReadUint()
	if err != nil {
		errorPage500(writer, params, "reading status failed: "+err.Error())
		return
	}

	switch status {
	case 0:
		kind, err := brReader.ReadUint()
		if err != nil {
			errorPage500(writer, params, "reading object kind failed: "+err.Error())
			return
		}

		switch kind {
		case 1:
			// Tree
			if redirectDir(writer, request) {
				return
			}
			count, err := brReader.ReadUint()
			if err != nil {
				errorPage500(writer, params, "reading entry count failed: "+err.Error())
				return
			}

			files := make([]displayTreeEntry, 0, count)
			for range count {
				typeCode, err := brReader.ReadUint()
				if err != nil {
					errorPage500(writer, params, "error reading entry type: "+err.Error())
					return
				}
				mode, err := brReader.ReadUint()
				if err != nil {
					errorPage500(writer, params, "error reading entry mode: "+err.Error())
					return
				}
				size, err := brReader.ReadUint()
				if err != nil {
					errorPage500(writer, params, "error reading entry size: "+err.Error())
					return
				}
				name, err := brReader.ReadData()
				if err != nil {
					errorPage500(writer, params, "error reading entry name: "+err.Error())
					return
				}
				files = append(files, displayTreeEntry{
					Name:      string(name),
					Mode:      fmt.Sprintf("%06o", mode),
					Size:      int64(size),
					IsFile:    typeCode == 2,
					IsSubtree: typeCode == 1,
				})
			}

			params["files"] = files
			params["readme_filename"] = "README.md"
			params["readme"] = template.HTML("<p>README rendering here is WIP again</p>") // TODO

			renderTemplate(writer, "repo_raw_dir", params)

		case 2:
			// Blob
			if redirectNoDir(writer, request) {
				return
			}
			content, err := brReader.ReadData()
			if err != nil && !errors.Is(err, io.EOF) {
				errorPage500(writer, params, "error reading blob content: "+err.Error())
				return
			}
			writer.Header().Set("Content-Type", "application/octet-stream")
			fmt.Fprint(writer, string(content))

		default:
			errorPage500(writer, params, fmt.Sprintf("unknown object kind: %d", kind))
		}

	case 3:
		errorPage500(writer, params, "path not found: "+pathSpec)

	default:
		errorPage500(writer, params, fmt.Sprintf("unknown status code: %d", status))
	}
}
