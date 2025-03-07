// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// query_name_desc_list is a helper function that executes a query and returns a
// list of name_desc_t results.
func query_name_desc_list(ctx context.Context, query string, args ...any) (result []name_desc_t, err error) {
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
		result = append(result, name_desc_t{name, description})
	}
	return result, rows.Err()
}

// name_desc_t holds a name and a description.
type name_desc_t struct {
	Name        string
	Description string
}
