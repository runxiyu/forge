// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/emersion/go-message"
	"github.com/jackc/pgx/v5"
)

// lmtpHandleMailingList stores an incoming email into the mailing list archive
// for the specified group/list. It expects the list to be already existing.
func (s *Server) lmtpHandleMailingList(session *lmtpSession, groupPath []string, listName string, email *message.Entity, raw []byte, envelopeFrom string) error {
	ctx := session.ctx

	groupID, err := s.resolveGroupPath(ctx, groupPath)
	if err != nil {
		return err
	}

	var listID int
	if err := s.database.QueryRow(ctx,
		`SELECT id FROM mailing_lists WHERE group_id = $1 AND name = $2`,
		groupID, listName,
	).Scan(&listID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("mailing list not found")
		}
		return err
	}

	title := email.Header.Get("Subject")
	sender := email.Header.Get("From")

	date := time.Now()
	if dh := email.Header.Get("Date"); dh != "" {
		if t, err := mail.ParseDate(dh); err == nil {
			date = t
		}
	}

	_, err = s.database.Exec(ctx, `INSERT INTO mailing_list_emails (list_id, title, sender, date, content) VALUES ($1, $2, $3, $4, $5)`, listID, title, sender, date, raw)
	if err != nil {
		return err
	}

	if derr := s.relayMailingListMessage(ctx, listID, envelopeFrom, raw); derr != nil {
		// for now, return the error to LMTP so the sender learns delivery failed...
		// should replace this with queueing or something nice
		return derr
	}

	return nil
}

// resolveGroupPath resolves a group path (segments) to a group ID.
func (s *Server) resolveGroupPath(ctx context.Context, groupPath []string) (int, error) {
	var groupID int
	err := s.database.QueryRow(ctx, `
	WITH RECURSIVE group_path_cte AS (
		SELECT id, parent_group, name, 1 AS depth
		FROM groups
		WHERE name = ($1::text[])[1]
			AND parent_group IS NULL

		UNION ALL

		SELECT g.id, g.parent_group, g.name, group_path_cte.depth + 1
		FROM groups g
		JOIN group_path_cte ON g.parent_group = group_path_cte.id
		WHERE g.name = ($1::text[])[group_path_cte.depth + 1]
			AND group_path_cte.depth + 1 <= cardinality($1::text[])
	)
	SELECT c.id
	FROM group_path_cte c
	WHERE c.depth = cardinality($1::text[])
	`, groupPath).Scan(&groupID)
	return groupID, err
}
