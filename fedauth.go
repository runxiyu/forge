// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
)

func check_and_update_federated_user_status(ctx context.Context, user_id int, service, remote_username, pubkey string) (bool, error) {
	var err error
	var resp *http.Response
	matched := false
	username_escaped := url.PathEscape(remote_username)
	switch service {
	case "sr.ht":
		resp, err = http.Get("https://meta.sr.ht/~" + username_escaped + ".keys")
	case "github":
		resp, err = http.Get("https://github.com/" + username_escaped + ".keys")
	case "codeberg":
		resp, err = http.Get("https://codeberg.org/" + username_escaped + ".keys")
	case "tangled":
		resp, err = http.Get("https://tangled.sh/keys/" + username_escaped)
		// TODO: Don't rely on one webview
	default:
		return false, errors.New("unknown federated service")
	}

	if err != nil {
		return false, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	buf := bufio.NewReader(resp.Body)

	for {
		line, err := buf.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return false, err
		}

		line_split := strings.Split(line, " ")
		if len(line_split) < 2 {
			continue
		}
		line = strings.Join(line_split[:2], " ")

		if line == pubkey {
			matched = true
			break
		}
	}

	if !matched {
		return false, nil
	}

	var tx pgx.Tx
	if tx, err = database.Begin(ctx); err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if _, err = tx.Exec(ctx, `UPDATE users SET type = 'federated' WHERE id = $1 AND type = 'pubkey_only'`, user_id); err != nil {
		return false, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO federated_identities (user_id, service, remote_username) VALUES ($1, $2, $3)`, user_id, service, remote_username); err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}
