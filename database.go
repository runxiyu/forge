// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// queryNameDesc is a helper function that executes a query and returns a
// list of name_desc_t results.
func queryNameDesc(ctx context.Context, query string, args ...any) (result []nameDesc, err error) {
	var rows pgx.Rows

	if rows, err = database.Query(ctx, query, args...); err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			return nil, err
		}
		result = append(result, nameDesc{name, description})
	}
	return result, rows.Err()
}

// nameDesc holds a name and a description.
type nameDesc struct {
	Name        string
	Description string
}
