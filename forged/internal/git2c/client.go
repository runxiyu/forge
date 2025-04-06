// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package git2c provides routines to interact with the git2d backend daemon.
package git2c

import (
	"fmt"
	"net"

	"git.sr.ht/~sircmpwn/go-bare"
)

// Client represents a connection to the git2d backend daemon.
type Client struct {
	socketPath string
	conn       net.Conn
	writer     *bare.Writer
	reader     *bare.Reader
}

// NewClient establishes a connection to a git2d socket and returns a new Client.
func NewClient(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("git2d connection failed: %w", err)
	}

	writer := bare.NewWriter(conn)
	reader := bare.NewReader(conn)

	return &Client{
		socketPath: socketPath,
		conn:       conn,
		writer:     writer,
		reader:     reader,
	}, nil
}

// Close terminates the underlying socket connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
