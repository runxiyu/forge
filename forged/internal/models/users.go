// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package models

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
)

// AddUserSSH adds a new user solely based on their SSH public key.
//
// TODO: Audit all users of this function.
func AddUserSSH(ctx context.Context, db database.Database, pubkey string) (userID int, err error) {
	var txn pgx.Tx

	if txn, err = db.Begin(ctx); err != nil {
		return
	}
	defer func() {
		_ = txn.Rollback(ctx)
	}()

	if err = txn.QueryRow(ctx, `INSERT INTO users (type) VALUES ('pubkey_only') RETURNING id`).Scan(&userID); err != nil {
		return
	}

	if _, err = txn.Exec(ctx, `INSERT INTO ssh_public_keys (key_string, user_id) VALUES ($1, $2)`, pubkey, userID); err != nil {
		return
	}

	err = txn.Commit(ctx)
	return
}
