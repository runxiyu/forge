// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/jackc/pgx/v5/pgtype"
)

func handle_repo_info(w http.ResponseWriter, r *http.Request, params map[string]any) (err error) {
	var group_path []string
	var repo_name, repo_path string

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

	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	w.WriteHeader(http.StatusOK)

	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", repo_path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer func() {
		_ = stdout.Close()
	}()
	cmd.Stderr = cmd.Stdout

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = pack_line(w, "# service=git-upload-pack\n"); err != nil {
		return err
	}

	if err = pack_flush(w); err != nil {
		return
	}

	if _, err = io.Copy(w, stdout); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// Taken from https://github.com/icyphox/legit, MIT license
func pack_line(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "%04x%s", len(s)+4, s)
	return err
}

// Taken from https://github.com/icyphox/legit, MIT license
func pack_flush(w io.Writer) error {
	_, err := fmt.Fprint(w, "0000")
	return err
}
