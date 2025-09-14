// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package unsorted

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	stdsmtp "net/smtp"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

// relayMailingListMessage connects to the configured SMTP relay and sends the
// raw message to all subscribers of the given list. The message is written verbatim
// from this point on and there is no modification of any headers or whatever,
func (s *Server) relayMailingListMessage(ctx context.Context, listID int, envelopeFrom string, raw []byte) error {
	rows, err := s.database.Query(ctx, `SELECT email FROM mailing_list_subscribers WHERE list_id = $1`, listID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var recipients []string
	for rows.Next() {
		var email string
		if err = rows.Scan(&email); err != nil {
			return err
		}
		recipients = append(recipients, email)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	if len(recipients) == 0 {
		slog.Info("mailing list has no subscribers", "list_id", listID)
		return nil
	}

	netw := s.config.SMTP.Net
	if netw == "" {
		netw = "tcp"
	}
	if s.config.SMTP.Addr == "" {
		return errors.New("smtp relay addr not configured")
	}
	helloName := s.config.SMTP.HelloName
	if helloName == "" {
		helloName = s.config.LMTP.Domain
	}
	transport := s.config.SMTP.Transport
	if transport == "" {
		transport = "plain"
	}

	switch transport {
	case "plain", "tls":
		d := net.Dialer{Timeout: 30 * time.Second}
		var conn net.Conn
		var err error
		if transport == "tls" {
			tlsCfg := &tls.Config{ServerName: hostFromAddr(s.config.SMTP.Addr), InsecureSkipVerify: s.config.SMTP.TLSInsecure}
			conn, err = tls.DialWithDialer(&d, netw, s.config.SMTP.Addr, tlsCfg)
		} else {
			conn, err = d.DialContext(ctx, netw, s.config.SMTP.Addr)
		}
		if err != nil {
			return fmt.Errorf("dial smtp: %w", err)
		}
		defer conn.Close()

		c := smtp.NewClient(conn)
		defer c.Close()

		if err := c.Hello(helloName); err != nil {
			return fmt.Errorf("smtp hello: %w", err)
		}

		if s.config.SMTP.Username != "" {
			mech := sasl.NewPlainClient("", s.config.SMTP.Username, s.config.SMTP.Password)
			if err := c.Auth(mech); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}

		if err := c.Mail(envelopeFrom, &smtp.MailOptions{}); err != nil {
			return fmt.Errorf("smtp mail from: %w", err)
		}
		for _, rcpt := range recipients {
			if err := c.Rcpt(rcpt, &smtp.RcptOptions{}); err != nil {
				return fmt.Errorf("smtp rcpt %s: %w", rcpt, err)
			}
		}
		wc, err := c.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		if _, err := wc.Write(raw); err != nil {
			_ = wc.Close()
			return fmt.Errorf("smtp write: %w", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("smtp data close: %w", err)
		}
		if err := c.Quit(); err != nil {
			return fmt.Errorf("smtp quit: %w", err)
		}
		return nil
	case "starttls":
		d := net.Dialer{Timeout: 30 * time.Second}
		conn, err := d.DialContext(ctx, netw, s.config.SMTP.Addr)
		if err != nil {
			return fmt.Errorf("dial smtp: %w", err)
		}
		defer conn.Close()

		host := hostFromAddr(s.config.SMTP.Addr)
		c, err := stdsmtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		defer c.Close()

		if err := c.Hello(helloName); err != nil {
			return fmt.Errorf("smtp hello: %w", err)
		}
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return errors.New("smtp server does not support STARTTLS")
		}
		tlsCfg := &tls.Config{ServerName: host, InsecureSkipVerify: s.config.SMTP.TLSInsecure} // #nosec G402
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}

		// seems like ehlo is required after starttls
		if err := c.Hello(helloName); err != nil {
			return fmt.Errorf("smtp hello (post-starttls): %w", err)
		}

		if s.config.SMTP.Username != "" {
			auth := stdsmtp.PlainAuth("", s.config.SMTP.Username, s.config.SMTP.Password, host)
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
		if err := c.Mail(envelopeFrom); err != nil {
			return fmt.Errorf("smtp mail from: %w", err)
		}
		for _, rcpt := range recipients {
			if err := c.Rcpt(rcpt); err != nil {
				return fmt.Errorf("smtp rcpt %s: %w", rcpt, err)
			}
		}
		wc, err := c.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		if _, err := wc.Write(raw); err != nil {
			_ = wc.Close()
			return fmt.Errorf("smtp write: %w", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("smtp data close: %w", err)
		}
		if err := c.Quit(); err != nil {
			return fmt.Errorf("smtp quit: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown smtp transport: %q", transport)
	}
}

func hostFromAddr(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host == "" {
		return addr
	}
	return host
}
