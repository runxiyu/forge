// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"encoding/hex"
	"fmt"
)

type TreeEntryRaw struct {
	Mode uint64
	Name string
	OID  string // hex
}

func (c *Client) TreeListByOID(repoPath, treeHex string) ([]TreeEntryRaw, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return nil, fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(9); err != nil {
		return nil, fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(treeHex)); err != nil {
		return nil, fmt.Errorf("sending tree oid failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return nil, fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return nil, Perror(status)
	}
	count, err := c.reader.ReadUint()
	if err != nil {
		return nil, fmt.Errorf("reading count failed: %w", err)
	}
	entries := make([]TreeEntryRaw, 0, count)
	for range count {
		mode, err := c.reader.ReadUint()
		if err != nil {
			return nil, fmt.Errorf("reading mode failed: %w", err)
		}
		name, err := c.reader.ReadData()
		if err != nil {
			return nil, fmt.Errorf("reading name failed: %w", err)
		}
		id, err := c.reader.ReadData()
		if err != nil {
			return nil, fmt.Errorf("reading oid failed: %w", err)
		}
		entries = append(entries, TreeEntryRaw{Mode: mode, Name: string(name), OID: hex.EncodeToString(id)})
	}
	return entries, nil
}

func (c *Client) WriteTree(repoPath string, entries []TreeEntryRaw) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(10); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteUint(uint64(len(entries))); err != nil {
		return "", fmt.Errorf("sending count failed: %w", err)
	}
	for _, e := range entries {
		if err := c.writer.WriteUint(e.Mode); err != nil {
			return "", fmt.Errorf("sending mode failed: %w", err)
		}
		if err := c.writer.WriteData([]byte(e.Name)); err != nil {
			return "", fmt.Errorf("sending name failed: %w", err)
		}
		raw, err := hex.DecodeString(e.OID)
		if err != nil {
			return "", fmt.Errorf("decode oid hex: %w", err)
		}
		if err := c.writer.WriteDataFixed(raw); err != nil {
			return "", fmt.Errorf("sending oid failed: %w", err)
		}
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return "", fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return "", Perror(status)
	}
	id, err := c.reader.ReadData()
	if err != nil {
		return "", fmt.Errorf("reading oid failed: %w", err)
	}
	return hex.EncodeToString(id), nil
}

func (c *Client) WriteBlob(repoPath string, content []byte) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(11); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData(content); err != nil {
		return "", fmt.Errorf("sending blob content failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return "", fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return "", Perror(status)
	}
	id, err := c.reader.ReadData()
	if err != nil {
		return "", fmt.Errorf("reading oid failed: %w", err)
	}
	return hex.EncodeToString(id), nil
}
