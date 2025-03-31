// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: All database handling logic in all request handlers must be revamped.
// We must ensure that each request has all logic in one transaction (subject
// to exceptions if appropriate) so they get a consistent view of the database
// at a single point. A failure to do so may cause things as serious as
// privilege escalation.

// database serves as the primary database handle for this entire application.
// Transactions or single reads may be used therefrom. A [pgxpool.Pool] is
// necessary to safely use pgx concurrently; pgx.Conn, etc. are insufficient.
var database *pgxpool.Pool

// queryNameDesc is a helper function that executes a query and returns a
// list of nameDesc results. The query must return two string arguments, i.e. a
// name and a description.
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
