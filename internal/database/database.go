// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package database provides stubs and wrappers for databases.
package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	*pgxpool.Pool
}

func Open(connString string) (Database, error) {
	db, err := pgxpool.New(context.Background(), connString)
	return Database{db}, err
}
