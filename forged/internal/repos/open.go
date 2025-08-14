// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package repos

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/gogit"
)

// Open opens a git repository by group and repo name.
//
// TODO: This should be deprecated in favor of doing it in the relevant
// request/router context in the future, as it cannot cover the nuance of
// fields needed.
func Open(ctx context.Context, db database.Database, groupPath []string, repoName string) (repo *gogit.Repository, description string, repoID int, fsPath string, err error) {
	err = db.QueryRow(ctx, `
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

	-- Recurse: join next segment of the path
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
SELECT
	r.filesystem_path,
	COALESCE(r.description, ''),
	r.id
FROM group_path_cte g
JOIN repos r ON r.group_id = g.id
WHERE g.depth = cardinality($1::text[])
	AND r.name = $2
	`, pgtype.FlatArray[string](groupPath), repoName).Scan(&fsPath, &description, &repoID)
	if err != nil {
		return
	}

	repo, err = gogit.Open(fsPath)
	return
}
