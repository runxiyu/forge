// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package irc provides basic IRC bot functionality.
package irc

import (
	"crypto/tls"
	"log/slog"
	"net"

	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

// Config contains IRC connection and identity settings for the bot.
// This should usually be a part of the primary config struct.
type Config struct {
	Net   string `scfg:"net"`
	Addr  string `scfg:"addr"`
	TLS   bool   `scfg:"tls"`
	SendQ uint   `scfg:"sendq"`
	Nick  string `scfg:"nick"`
	User  string `scfg:"user"`
	Gecos string `scfg:"gecos"`
}

// Bot represents an IRC bot client that handles events and allows for sending messages.
type Bot struct {
	config            *Config
	ircSendBuffered   chan string
	ircSendDirectChan chan misc.ErrorBack[string]
}

// NewBot creates a new Bot instance using the provided configuration.
func NewBot(c *Config) (b *Bot) {
	b = &Bot{
		config: c,
	}
	return
}

// Connect establishes a new IRC session and starts handling incoming and outgoing messages.
// This method blocks until an error occurs or the connection is closed.
func (b *Bot) Connect() error {
	var err error
	var underlyingConn net.Conn
	if b.config.TLS {
		underlyingConn, err = tls.Dial(b.config.Net, b.config.Addr, nil)
	} else {
		underlyingConn, err = net.Dial(b.config.Net, b.config.Addr)
	}
	if err != nil {
		return err
	}
	defer underlyingConn.Close()

	conn := NewConn(underlyingConn)

	logAndWriteLn := func(s string) (n int, err error) {
		slog.Debug("irc tx", "line", s)
		return conn.WriteString(s + "\r\n")
	}

	_, err = logAndWriteLn("NICK " + b.config.Nick)
	if err != nil {
		return err
	}
	_, err = logAndWriteLn("USER " + b.config.User + " 0 * :" + b.config.Gecos)
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
				c, ok := msg.Source.(Client)
				if !ok {
					slog.Error("unable to convert source of JOIN to client")
				}
				if c.Nick != b.config.Nick {
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
		case line := <-b.ircSendBuffered:
			_, err = logAndWriteLn(line)
			if err != nil {
				select {
				case b.ircSendBuffered <- line:
				default:
					slog.Error("unable to requeue message", "line", line)
				}
				writeLoopAbort <- struct{}{}
				return err
			}
		case lineErrorBack := <-b.ircSendDirectChan:
			_, err = logAndWriteLn(lineErrorBack.Content)
			lineErrorBack.ErrorChan <- err
			if err != nil {
				writeLoopAbort <- struct{}{}
				return err
			}
		}
	}
}

// SendDirect sends an IRC message directly to the connection and bypasses
// the buffering system.
func (b *Bot) SendDirect(line string) error {
	ech := make(chan error, 1)

	b.ircSendDirectChan <- misc.ErrorBack[string]{
		Content:   line,
		ErrorChan: ech,
	}

	return <-ech
}

// Send queues a message to be sent asynchronously via the buffered send queue.
// If the queue is full, the message is dropped and an error is logged.
func (b *Bot) Send(line string) {
	select {
	case b.ircSendBuffered <- line:
	default:
		slog.Error("irc sendq full", "line", line)
	}
}

// ConnectLoop continuously attempts to maintain an IRC session.
// If the connection drops, it automatically retries with no delay.
func (b *Bot) ConnectLoop() {
	b.ircSendBuffered = make(chan string, b.config.SendQ)
	b.ircSendDirectChan = make(chan misc.ErrorBack[string])

	for {
		err := b.Connect()
		slog.Error("irc session error", "error", err)
	}
}
