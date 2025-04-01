// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
// SPDX-FileCopyrightText: Copyright (c) 2024 Robin Jarry <robin@jarry.cc>

package main

import (
	"bytes"
	"errors"
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
	from string
	to   []string
}

func (session *lmtpSession) Reset() {
	session.from = ""
	session.to = nil
}

func (session *lmtpSession) Logout() error {
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
	// TODO
	session := &lmtpSession{}
	return session, nil
}

func serveLMTP(listener net.Listener) error {
	// TODO: Manually construct smtp.Server
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
		// disregard automatic emails like OOO replies
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
	// still be valid when the queued task function is evaluated.
	from = session.from
	to = session.to

	// TODO: Process the actual message contents
	_, _ = from, to

end:
	session.to = nil
	session.from = ""
	return err
}
