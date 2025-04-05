// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"crypto/tls"
	"log/slog"
	"net"

	irc "go.lindenii.runxiyu.org/lindenii-irc"
)

var (
	ircSendBuffered   chan string
	ircSendDirectChan chan errorBack[string]
)

type errorBack[T any] struct {
	content   T
	errorBack chan error
}

func (s *server) ircBotSession() error {
	var err error
	var underlyingConn net.Conn
	if s.config.IRC.TLS {
		underlyingConn, err = tls.Dial(s.config.IRC.Net, s.config.IRC.Addr, nil)
	} else {
		underlyingConn, err = net.Dial(s.config.IRC.Net, s.config.IRC.Addr)
	}
	if err != nil {
		return err
	}
	defer underlyingConn.Close()

	conn := irc.NewConn(underlyingConn)

	logAndWriteLn := func(s string) (n int, err error) {
		slog.Debug("irc tx", "line", s)
		return conn.WriteString(s + "\r\n")
	}

	_, err = logAndWriteLn("NICK " + s.config.IRC.Nick)
	if err != nil {
		return err
	}
	_, err = logAndWriteLn("USER " + s.config.IRC.User + " 0 * :" + s.config.IRC.Gecos)
	if err != nil {
		return err
	}

	readLoopError := make(chan error)
	writeLoopAbort := make(chan struct{})
	go func() {
		for {
			select {
			case <-writeLoopAbort:
				return
			default:
			}

			msg, line, err := conn.ReadMessage()
			if err != nil {
				readLoopError <- err
				return
			}

			slog.Debug("irc rx", "line", line)

			switch msg.Command {
			case "001":
				_, err = logAndWriteLn("JOIN #chat")
				if err != nil {
					readLoopError <- err
					return
				}
			case "PING":
				_, err = logAndWriteLn("PONG :" + msg.Args[0])
				if err != nil {
					readLoopError <- err
					return
				}
			case "JOIN":
				c, ok := msg.Source.(irc.Client)
				if !ok {
					slog.Error("unable to convert source of JOIN to client")
				}
				if c.Nick != s.config.IRC.Nick {
					continue
				}
			default:
			}
		}
	}()

	for {
		select {
		case err = <-readLoopError:
			return err
		case line := <-ircSendBuffered:
			_, err = logAndWriteLn(line)
			if err != nil {
				select {
				case ircSendBuffered <- line:
				default:
					slog.Error("unable to requeue message", "line", line)
				}
				writeLoopAbort <- struct{}{}
				return err
			}
		case lineErrorBack := <-ircSendDirectChan:
			_, err = logAndWriteLn(lineErrorBack.content)
			lineErrorBack.errorBack <- err
			if err != nil {
				writeLoopAbort <- struct{}{}
				return err
			}
		}
	}
}

// ircSendDirect sends an IRC message directly to the connection and bypasses
// the buffering system.
func ircSendDirect(s string) error {
	ech := make(chan error, 1)

	ircSendDirectChan <- errorBack[string]{
		content:   s,
		errorBack: ech,
	}

	return <-ech
}

// TODO: Delay and warnings?
func (s *server) ircBotLoop() {
	ircSendBuffered = make(chan string, s.config.IRC.SendQ)
	ircSendDirectChan = make(chan errorBack[string])

	for {
		err := s.ircBotSession()
		slog.Error("irc session error", "error", err)
	}
}
