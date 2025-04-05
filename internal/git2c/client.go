// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"fmt"
	"net"

	"git.sr.ht/~sircmpwn/go-bare"
)

type Client struct {
	SocketPath string
	conn       net.Conn
	writer     *bare.Writer
	reader     *bare.Reader
}

func NewClient(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("git2d connection failed: %w", err)
	}

	writer := bare.NewWriter(conn)
	reader := bare.NewReader(conn)

	return &Client{
		SocketPath: socketPath,
		conn:       conn,
		writer:     writer,
		reader:     reader,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
