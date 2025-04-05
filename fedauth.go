// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

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

// fedauth checks whether a user's SSH public key matches the remote username
// they claim to have on the service. If so, the association is recorded.
func (s *Server) fedauth(ctx context.Context, userID int, service, remoteUsername, pubkey string) (bool, error) {
	var err error

	matched := false
	usernameEscaped := url.PathEscape(remoteUsername)

	var req *http.Request
	switch service {
	// TODO: Services should be configurable by the instance administrator
	// and should not be hardcoded in the source code.
	case "sr.ht":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://meta.sr.ht/~"+usernameEscaped+".keys", nil)
	case "github":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://github.com/"+usernameEscaped+".keys", nil)
	case "codeberg":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://codeberg.org/"+usernameEscaped+".keys", nil)
	case "tangled":
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, "https://tangled.sh/keys/"+usernameEscaped, nil)
		// TODO: Don't rely on one webview
	default:
		return false, errors.New("unknown federated service")
	}
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
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

		lineSplit := strings.Split(line, " ")
		if len(lineSplit) < 2 {
			continue
		}
		line = strings.Join(lineSplit[:2], " ")

		if line == pubkey {
			matched = true
			break
		}
	}

	if !matched {
		return false, nil
	}

	var txn pgx.Tx
	if txn, err = s.Database.Begin(ctx); err != nil {
		return false, err
	}
	defer func() {
		_ = txn.Rollback(ctx)
	}()
	if _, err = txn.Exec(ctx, `UPDATE users SET type = 'federated' WHERE id = $1 AND type = 'pubkey_only'`, userID); err != nil {
		return false, err
	}
	if _, err = txn.Exec(ctx, `INSERT INTO federated_identities (user_id, service, remote_username) VALUES ($1, $2, $3)`, userID, service, remoteUsername); err != nil {
		return false, err
	}
	if err = txn.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}
