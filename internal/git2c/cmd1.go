// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

func (c *Client) Cmd1(repoPath string) ([]Commit, *FilenameContents, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return nil, nil, fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(1); err != nil {
		return nil, nil, fmt.Errorf("sending command failed: %w", err)
	}

	status, err := c.reader.ReadUint()
	if err != nil {
		return nil, nil, fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return nil, nil, fmt.Errorf("git2d error: %d", status)
	}

	// README
	readmeRaw, err := c.reader.ReadData()
	if err != nil {
		readmeRaw = nil
	}

	readmeFilename := "README.md" // TODO
	readme := &FilenameContents{Filename: readmeFilename, Content: readmeRaw}

	// Commits
	var commits []Commit
	for {
		id, err := c.reader.ReadData()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, nil, fmt.Errorf("reading commit ID failed: %w", err)
		}
		title, _ := c.reader.ReadData()
		authorName, _ := c.reader.ReadData()
		authorEmail, _ := c.reader.ReadData()
		authorDate, _ := c.reader.ReadData()

		commits = append(commits, Commit{
			Hash:    hex.EncodeToString(id),
			Author:  string(authorName),
			Email:   string(authorEmail),
			Date:    string(authorDate),
			Message: string(title),
		})
	}

	return commits, readme, nil
}
