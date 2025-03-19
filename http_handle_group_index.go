// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func httpHandleGroupIndex(w http.ResponseWriter, r *http.Request, params map[string]any) {
	var groupPath []string
	var repos []nameDesc
	var subgroups []nameDesc
	var err error
	var groupID int
	var groupDesc string

	groupPath = params["group_path"].([]string)

	// The group itself
	err = database.QueryRow(r.Context(), `
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

	if err == pgx.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error getting group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ACL
	var count int
	err = database.QueryRow(r.Context(), `
		SELECT COUNT(*)
		FROM user_group_roles
		WHERE user_id = $1
			AND group_id = $2
	`, params["user_id"].(int), groupID).Scan(&count)
	if err != nil {
		http.Error(w, "Error checking access: "+err.Error(), http.StatusInternalServerError)
		return
	}
	directAccess := (count > 0)

	if r.Method == "POST" {
		if !directAccess {
			http.Error(w, "You do not have direct access to this group", http.StatusForbidden)
			return
		}

		repoName := r.FormValue("repo_name")
		repoDesc := r.FormValue("repo_desc")
		contribReq := r.FormValue("repo_contrib")
		if repoName == "" {
			http.Error(w, "Repo name is required", http.StatusBadRequest)
			return
		}

		var newRepoID int
		err := database.QueryRow(
			r.Context(),
			`INSERT INTO repos (name, description, group_id, contrib_requirements)
	 VALUES ($1, $2, $3, $4)
	 RETURNING id`,
			repoName,
			repoDesc,
			groupID,
			contribReq,
		).Scan(&newRepoID)
		if err != nil {
			http.Error(w, "Error creating repo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filePath := filepath.Join(config.Git.RepoDir, strconv.Itoa(newRepoID)+".git")

		_, err = database.Exec(
			r.Context(),
			`UPDATE repos
	 SET filesystem_path = $1
	 WHERE id = $2`,
			filePath,
			newRepoID,
		)
		if err != nil {
			http.Error(w, "Error updating repo path: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err = gitInit(filePath); err != nil {
			http.Error(w, "Error initializing repo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		redirectUnconditionally(w, r)
		return
	}

	// Repos
	var rows pgx.Rows
	rows, err = database.Query(r.Context(), `
		SELECT name, COALESCE(description, '')
		FROM repos
		WHERE group_id = $1
	`, groupID)
	if err != nil {
		http.Error(w, "Error getting repos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			http.Error(w, "Error getting repos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		repos = append(repos, nameDesc{name, description})
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Error getting repos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Subgroups
	rows, err = database.Query(r.Context(), `
		SELECT name, COALESCE(description, '')
		FROM groups
		WHERE parent_group = $1
	`, groupID)
	if err != nil {
		http.Error(w, "Error getting subgroups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			http.Error(w, "Error getting subgroups: "+err.Error(), http.StatusInternalServerError)
			return
		}
		subgroups = append(subgroups, nameDesc{name, description})
	}
	if err = rows.Err(); err != nil {
		http.Error(w, "Error getting subgroups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	params["repos"] = repos
	params["subgroups"] = subgroups
	params["description"] = groupDesc
	params["direct_access"] = directAccess

	renderTemplate(w, "group", params)
}
