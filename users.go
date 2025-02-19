package main

import (
	"context"
)

func add_user_ssh(ctx context.Context, pubkey string) (user_id int, err error) {
	tx, err := database.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `INSERT INTO users (type) VALUES ('pubkey_only') RETURNING id`).Scan(&user_id)
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx, `INSERT INTO ssh_public_keys (key_string, user_id) VALUES ($1, $2)`, pubkey, user_id)
	if err != nil {
		return
	}

	err = tx.Commit(ctx)
	return
}
