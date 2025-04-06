// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package database provides stubs and wrappers for databases.
package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database is a wrapper around pgxpool.Pool to provide a common interface for
// other packages in the forge.
type Database struct {
	*pgxpool.Pool
}

// Open opens a new database connection pool using the provided connection
// string. It returns a Database instance and an error if any occurs.
// It is run indefinitely in the background.
func Open(connString string) (Database, error) {
	db, err := pgxpool.New(context.Background(), connString)
	return Database{db}, err
}
