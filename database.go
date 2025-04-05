// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TODO: All database handling logic in all request handlers must be revamped.
// We must ensure that each request has all logic in one transaction (subject
// to exceptions if appropriate) so they get a consistent view of the database
// at a single point. A failure to do so may cause things as serious as
// privilege escalation.

// QueryNameDesc is a helper function that executes a query and returns a
// list of nameDesc results. The query must return two string arguments, i.e. a
// name and a description.
func (s *Server) QueryNameDesc(ctx context.Context, query string, args ...any) (result []NameDesc, err error) {
	var rows pgx.Rows

	if rows, err = s.database.Query(ctx, query, args...); err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		if err = rows.Scan(&name, &description); err != nil {
			return nil, err
		}
		result = append(result, NameDesc{name, description})
	}
	return result, rows.Err()
}

// NameDesc holds a name and a description.
type NameDesc struct {
	Name        string
	Description string
}
