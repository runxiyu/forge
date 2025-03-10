// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/jackc/pgx/v5/pgtype"
)

func handle_upload_pack(w http.ResponseWriter, r *http.Request, params map[string]any) (err error) {
	var group_path []string
	var repo_name string
	var repo_path string
	var stdout io.ReadCloser
	var stdin io.WriteCloser
	var cmd *exec.Cmd

	group_path, repo_name = params["group_path"].([]string), params["repo_name"].(string)

	if err := database.QueryRow(r.Context(), `
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
		pgtype.FlatArray[string](group_path),
		repo_name,
	).Scan(&repo_path); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	cmd = exec.Command("git", "upload-pack", "--stateless-rpc", repo_path)
	cmd.Env = append(os.Environ(), "LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket)
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	defer func() {
		_ = stdout.Close()
	}()

	if stdin, err = cmd.StdinPipe(); err != nil {
		return err
	}
	defer func() {
		_ = stdin.Close()
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	if _, err = io.Copy(stdin, r.Body); err != nil {
		return err
	}

	if err = stdin.Close(); err != nil {
		return err
	}

	if _, err = io.Copy(w, stdout); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}
