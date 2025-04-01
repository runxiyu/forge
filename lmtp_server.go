// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
// SPDX-FileCopyrightText: Copyright (c) 2024 Robin Jarry <robin@jarry.cc>

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-smtp"
)

type lmtpHandler struct{}

type lmtpSession struct {
	from   string
	to     []string
	ctx    context.Context
	cancel context.CancelFunc
}

func (session *lmtpSession) Reset() {
	session.from = ""
	session.to = nil
}

func (session *lmtpSession) Logout() error {
	session.cancel()
	return nil
}

func (session *lmtpSession) AuthPlain(_, _ string) error {
	return nil
}

func (session *lmtpSession) Mail(from string, _ *smtp.MailOptions) error {
	session.from = from
	return nil
}

func (session *lmtpSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	session.to = append(session.to, to)
	return nil
}

func (*lmtpHandler) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	ctx, cancel := context.WithCancel(context.Background())
	session := &lmtpSession{
		ctx:    ctx,
		cancel: cancel,
	}
	return session, nil
}

func serveLMTP(listener net.Listener) error {
	smtpServer := smtp.NewServer(&lmtpHandler{})
	smtpServer.LMTP = true
	smtpServer.Domain = config.LMTP.Domain
	smtpServer.Addr = config.LMTP.Socket
	smtpServer.WriteTimeout = time.Duration(config.LMTP.WriteTimeout) * time.Second
	smtpServer.ReadTimeout = time.Duration(config.LMTP.ReadTimeout) * time.Second
	smtpServer.EnableSMTPUTF8 = true
	return smtpServer.Serve(listener)
}

func (session *lmtpSession) Data(r io.Reader) error {
	var (
		email *message.Entity
		from  string
		to    []string
		err   error
		buf   bytes.Buffer
		data  []byte
		n     int64
	)

	n, err = io.CopyN(&buf, r, config.LMTP.MaxSize)
	switch {
	case n == config.LMTP.MaxSize:
		err = errors.New("Message too big.")
		// drain whatever is left in the pipe
		_, _ = io.Copy(io.Discard, r)
		goto end
	case errors.Is(err, io.EOF):
		// message was smaller than max size
		break
	case err != nil:
		goto end
	}

	data = buf.Bytes()

	email, err = message.Read(bytes.NewReader(data))
	if err != nil && message.IsUnknownCharset(err) {
		goto end
	}

	switch strings.ToLower(email.Header.Get("Auto-Submitted")) {
	case "auto-generated", "auto-replied":
		// Disregard automatic emails like OOO replies.
		slog.Info("ignoring automatic message",
			"from", session.from,
			"to", strings.Join(session.to, ","),
			"message-id", email.Header.Get("Message-Id"),
			"subject", email.Header.Get("Subject"),
		)
		goto end
	}

	slog.Info("message received",
		"from", session.from,
		"to", strings.Join(session.to, ","),
		"message-id", email.Header.Get("Message-Id"),
		"subject", email.Header.Get("Subject"),
	)

	// Make local copies of the values before to ensure the references will
	// still be valid when the task is run.
	from = session.from
	to = session.to

	_ = from

	for _, to := range to {
		if !strings.HasSuffix(to, "@"+config.LMTP.Domain) {
			continue
		}
		localPart := to[:len(to)-len("@"+config.LMTP.Domain)]
		var segments []string
		segments, err = pathToSegments(localPart)
		if err != nil {
			// TODO: Should the entire email fail or should we just
			// notify them out of band?
			err = fmt.Errorf("cannot parse path: %w", err)
			goto end
		}
		sepIndex := -1
		for i, part := range segments {
			if part == "-" {
				sepIndex = i
				break
			}
		}
		if segments[len(segments)-1] == "" {
			segments = segments[:len(segments)-1] // We don't care about dir or not.
		}
		if sepIndex == -1 || len(segments) <= sepIndex+2 {
			err = errors.New("illegal path")
			goto end
		}

		groupPath := segments[:sepIndex]
		moduleType := segments[sepIndex+1]
		moduleName := segments[sepIndex+2]
		switch moduleType {
		case "repos":
			err = lmtpHandlePatch(groupPath, moduleName, email)
			if err != nil {
				goto end
			}
		default:
			err = errors.New("Emailing any endpoint other than repositories, is not supported yet.") // TODO
			goto end
		}
	}

end:
	session.to = nil
	session.from = ""
	switch err {
	case nil:
		return nil
	default:
		return &smtp.SMTPError{
			Code:    550,
			Message: err.Error(),
		}
	}
}
