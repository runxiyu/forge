// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import "fmt"

func (c *Client) InitRepo(repoPath, hooksPath string) error {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(15); err != nil {
		return fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(hooksPath)); err != nil {
		return fmt.Errorf("sending hooks path failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return Perror(status)
	}
	return nil
}
