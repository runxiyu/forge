// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"errors"
	"fmt"
	"io"
)

// CmdTreeRaw queries git2d for a tree or blob object at the given path within the repository.
// It returns either a directory listing or the contents of a file.
func (c *Client) CmdTreeRaw(repoPath, pathSpec string) ([]TreeEntry, string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return nil, "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(2); err != nil {
		return nil, "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(pathSpec)); err != nil {
		return nil, "", fmt.Errorf("sending path failed: %w", err)
	}

	status, err := c.reader.ReadUint()
	if err != nil {
		return nil, "", fmt.Errorf("reading status failed: %w", err)
	}

	switch status {
	case 0:
		kind, err := c.reader.ReadUint()
		if err != nil {
			return nil, "", fmt.Errorf("reading object kind failed: %w", err)
		}

		switch kind {
		case 1:
			// Tree
			count, err := c.reader.ReadUint()
			if err != nil {
				return nil, "", fmt.Errorf("reading entry count failed: %w", err)
			}

			var files []TreeEntry
			for range count {
				typeCode, err := c.reader.ReadUint()
				if err != nil {
					return nil, "", fmt.Errorf("error reading entry type: %w", err)
				}
				mode, err := c.reader.ReadUint()
				if err != nil {
					return nil, "", fmt.Errorf("error reading entry mode: %w", err)
				}
				size, err := c.reader.ReadUint()
				if err != nil {
					return nil, "", fmt.Errorf("error reading entry size: %w", err)
				}
				name, err := c.reader.ReadData()
				if err != nil {
					return nil, "", fmt.Errorf("error reading entry name: %w", err)
				}

				files = append(files, TreeEntry{
					Name:      string(name),
					Mode:      fmt.Sprintf("%06o", mode),
					Size:      size,
					IsFile:    typeCode == 2,
					IsSubtree: typeCode == 1,
				})
			}

			return files, "", nil

		case 2:
			// Blob
			content, err := c.reader.ReadData()
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, "", fmt.Errorf("error reading file content: %w", err)
			}

			return nil, string(content), nil

		default:
			return nil, "", fmt.Errorf("unknown kind: %d", kind)
		}

	case 3:
		return nil, "", fmt.Errorf("path not found: %s", pathSpec)

	default:
		return nil, "", fmt.Errorf("unknown status code: %d", status)
	}
}
