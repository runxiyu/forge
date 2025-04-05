// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"crypto/tls"
	"log/slog"
	"net"

	irc "go.lindenii.runxiyu.org/lindenii-irc"
)

type errorBack[T any] struct {
	content   T
	errorBack chan error
}

func (s *Server) ircBotSession() error {
	var err error
	var underlyingConn net.Conn
	if s.Config.IRC.TLS {
		underlyingConn, err = tls.Dial(s.Config.IRC.Net, s.Config.IRC.Addr, nil)
	} else {
		underlyingConn, err = net.Dial(s.Config.IRC.Net, s.Config.IRC.Addr)
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

	_, err = logAndWriteLn("NICK " + s.Config.IRC.Nick)
	if err != nil {
		return err
	}
	_, err = logAndWriteLn("USER " + s.Config.IRC.User + " 0 * :" + s.Config.IRC.Gecos)
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
				if c.Nick != s.Config.IRC.Nick {
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
		case line := <-s.IrcSendBuffered:
			_, err = logAndWriteLn(line)
			if err != nil {
				select {
				case s.IrcSendBuffered <- line:
				default:
					slog.Error("unable to requeue message", "line", line)
				}
				writeLoopAbort <- struct{}{}
				return err
			}
		case lineErrorBack := <-s.IrcSendDirectChan:
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
func (s *Server) ircSendDirect(line string) error {
	ech := make(chan error, 1)

	s.IrcSendDirectChan <- errorBack[string]{
		content:   line,
		errorBack: ech,
	}

	return <-ech
}

// TODO: Delay and warnings?
func (s *Server) ircBotLoop() {
	s.IrcSendBuffered = make(chan string, s.Config.IRC.SendQ)
	s.IrcSendDirectChan = make(chan errorBack[string])

	for {
		err := s.ircBotSession()
		slog.Error("irc session error", "error", err)
	}
}
