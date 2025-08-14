// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package web

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"go.lindenii.runxiyu.org/forge/forged/internal/gitcmd"
)

// httpHandleUploadPack handles incoming Git fetch/pull/clone's over the Smart
// HTTP protocol.
func (s *Server) httpHandleUploadPack(writer http.ResponseWriter, request *http.Request, params map[string]any) (err error) {
	if ct := request.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/x-git-upload-pack-request") {
		http.Error(writer, "bad content-type", http.StatusUnsupportedMediaType)
		return nil
	}

	decoded, err := decodeBody(request)
	if err != nil {
		http.Error(writer, "cannot decode request body", http.StatusBadRequest)
		return err
	}
	defer decoded.Close()

	var groupPath []string
	var repoName string
	var repoPath string

	groupPath, repoName = params["group_path"].([]string), params["repo_name"].(string)

	if err := s.database.QueryRow(request.Context(), `
	WITH RECURSIVE group_path_cte AS (
		-- Start: match the first name in the path where parent_group IS NULL
		SELECT
			id,
			parent_group,
			name,
			1 AS depth
		FROM groups
		WHERE name = ($1::text[])[1]
			AND parent_group IS NULL
	
		UNION ALL
	
		-- Recurse: jion next segment of the path
		SELECT
			g.id,
			g.parent_group,
			g.name,
			group_path_cte.depth + 1
		FROM groups g
		JOIN group_path_cte ON g.parent_group = group_path_cte.id
		WHERE g.name = ($1::text[])[group_path_cte.depth + 1]
			AND group_path_cte.depth + 1 <= cardinality($1::text[])
	)
	SELECT r.filesystem_path
	FROM group_path_cte c
	JOIN repos r ON r.group_id = c.id
	WHERE c.depth = cardinality($1::text[])
		AND r.name = $2
	`,
		pgtype.FlatArray[string](groupPath),
		repoName,
	).Scan(&repoPath); err != nil {
		return err
	}

	writer.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	// writer.Header().Set("Connection", "Keep-Alive")
	// writer.Header().Set("Transfer-Encoding", "chunked")

	env := []string{"LINDENII_FORGE_HOOKS_SOCKET_PATH=" + s.config.Hooks.Socket}
	if gp := request.Header.Get("Git-Protocol"); gp != "" {
		env = append(env, "GIT_PROTOCOL="+gp)
	}
	var stderrBuf bytes.Buffer
	err = gitcmd.UploadPack(request.Context(), repoPath, env, decoded, writer, &stderrBuf, "--stateless-rpc")
	if err != nil {
		log.Println(stderrBuf.String())
	}
	return err
}

func decodeBody(r *http.Request) (io.ReadCloser, error) {
	switch ce := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Encoding"))); ce {
	case "", "identity":
		return r.Body, nil
	case "gzip":
		zr, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		return zr, nil
	case "deflate":
		zr, err := zlib.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		return zr, nil
	default:
		return nil, fmt.Errorf("unsupported Content-Encoding: %q", ce)
	}
}
