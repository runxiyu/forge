// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func add_user_ssh(ctx context.Context, pubkey string) (user_id int, err error) {
	var tx pgx.Tx

	if tx, err = database.Begin(ctx); err != nil {
		return
	}
	defer tx.Rollback(ctx)

	if err = tx.QueryRow(ctx, `INSERT INTO users (type) VALUES ('pubkey_only') RETURNING id`).Scan(&user_id); err != nil {
		return
	}

	if _, err = tx.Exec(ctx, `INSERT INTO ssh_public_keys (key_string, user_id) VALUES ($1, $2)`, pubkey, user_id); err != nil {
		return
	}

	err = tx.Commit(ctx)
	return
}
