// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package database provides stubs and wrappers for databases.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	*pgxpool.Pool
}

func Open(ctx context.Context, conn string) (Database, error) {
	db, err := pgxpool.New(ctx, conn)
	if err != nil {
		err = fmt.Errorf("create pgxpool: %w", err)
	}
	return Database{db}, err
}
