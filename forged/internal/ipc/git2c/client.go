// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"context"
	"fmt"
	"net"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/bare"
)

// Client represents a connection to the git2d backend daemon.
type Client struct {
	socketPath string
	conn       net.Conn
	writer     *bare.Writer
	reader     *bare.Reader
}

// NewClient establishes a connection to a git2d socket and returns a new Client.
func NewClient(ctx context.Context, socketPath string) (*Client, error) {
	dialer := &net.Dialer{} //exhaustruct:ignore
	conn, err := dialer.DialContext(ctx, "unix", socketPath)
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
func (c *Client) Close() (err error) {
	if c.conn != nil {
		err = c.conn.Close()
		if err != nil {
			return fmt.Errorf("close underlying socket: %w", err)
		}
	}
	return nil
}
