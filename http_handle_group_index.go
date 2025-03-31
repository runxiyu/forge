// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func httpHandleGroupIndex(writer http.ResponseWriter, request *http.Request, params map[string]any) {
	var groupPath []string
	var repos []nameDesc
	var subgroups []nameDesc
	var err error
	var groupID int
	var groupDesc string

	groupPath = params["group_path"].([]string)

	// The group itself
	err = database.QueryRow(request.Context(), `
		WITH RECURSIVE group_path_cte AS (
			SELECT
				id,
				parent_group,
				name,
				1 AS depth
			FROM groups
			WHERE name = ($1::text[])[1]
				AND parent_group IS NULL

			UNION ALL

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
		SELECT c.id, COALESCE(g.description, '')
		FROM group_path_cte c
		JOIN groups g ON g.id = c.id
		WHERE c.depth = cardinality($1::text[])
	`,
		pgtype.FlatArray[string](groupPath),
	).Scan(&groupID, &groupDesc)

	if errors.Is(err, pgx.ErrNoRows) {
		errorPage404(writer, params)
		return
	} else if err != nil {
		errorPage500(writer, params, "Error getting group: "+err.Error())
		return
	}

	// ACL
	var count int
	err = database.QueryRow(request.Context(), `
		SELECT COUNT(*)
		FROM user_group_roles
		WHERE user_id = $1
			AND group_id = $2
	`, params["user_id"].(int), groupID).Scan(&count)
	if err != nil {
		errorPage500(writer, params, "Error checking access: "+err.Error())
		return
	}
	directAccess := (count > 0)

	if request.Method == http.MethodPost {
		if !directAccess {
			errorPage403(writer, params, "You do not have direct access to this group")
			return
		}

		repoName := request.FormValue("repo_name")
		repoDesc := request.FormValue("repo_desc")
		contribReq := request.FormValue("repo_contrib")
		if repoName == "" {
			errorPage400(writer, params, "Repo name is required")
			return
		}

		var newRepoID int
		err := database.QueryRow(
			request.Context(),
			`INSERT INTO repos (name, description, group_id, contrib_requirements)
	 VALUES ($1, $2, $3, $4)
	 RETURNING id`,
			repoName,
			repoDesc,
			groupID,
			contribReq,
		).Scan(&newRepoID)
		if err != nil {
			errorPage500(writer, params, "Error creating repo: "+err.Error())
			return
		}

		filePath := filepath.Join(config.Git.RepoDir, strconv.Itoa(newRepoID)+".git")

		_, err = database.Exec(
			request.Context(),
			`UPDATE repos
	 SET filesystem_path = $1
	 WHERE id = $2`,
			filePath,
			newRepoID,
		)
		if err != nil {
			errorPage500(writer, params, "Error updating repo path: "+err.Error())
			return
		}

		if err = gitInit(filePath); err != nil {
			errorPage500(writer, params, "Error initializing repo: "+err.Error())
			return
		}

		redirectUnconditionally(writer, request)
		return
	}

	// Repos
	var rows pgx.Rows
	rows, err = database.Query(request.Context(), `
		SELECT name, COALESCE(description, '')
		FROM repos
		WHERE group_id = $1
	`, groupID)
	if err != nil {
		errorPage500(writer, params, "Error getting repos: "+err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			errorPage500(writer, params, "Error getting repos: "+err.Error())
			return
		}
		repos = append(repos, nameDesc{name, description})
	}
	if err = rows.Err(); err != nil {
		errorPage500(writer, params, "Error getting repos: "+err.Error())
		return
	}

	// Subgroups
	rows, err = database.Query(request.Context(), `
		SELECT name, COALESCE(description, '')
		FROM groups
		WHERE parent_group = $1
	`, groupID)
	if err != nil {
		errorPage500(writer, params, "Error getting subgroups: "+err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			errorPage500(writer, params, "Error getting subgroups: "+err.Error())
			return
		}
		subgroups = append(subgroups, nameDesc{name, description})
	}
	if err = rows.Err(); err != nil {
		errorPage500(writer, params, "Error getting subgroups: "+err.Error())
		return
	}

	params["repos"] = repos
	params["subgroups"] = subgroups
	params["description"] = groupDesc
	params["direct_access"] = directAccess

	renderTemplate(writer, "group", params)
}
