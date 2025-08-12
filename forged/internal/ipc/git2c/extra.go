// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"encoding/hex"
	"fmt"
	"time"
)

type DiffChunk struct {
	Op      uint64
	Content string
}

type FileDiff struct {
	FromMode uint64
	ToMode   uint64
	FromPath string
	ToPath   string
	Chunks   []DiffChunk
}

type CommitInfo struct {
	Hash           string
	AuthorName     string
	AuthorEmail    string
	AuthorWhen     int64 // unix secs
	AuthorTZMin    int64 // minutes ofs
	CommitterName  string
	CommitterEmail string
	CommitterWhen  int64
	CommitterTZMin int64
	Message        string
	Parents        []string // hex
	Files          []FileDiff
}

func (c *Client) ResolveRef(repoPath, refType, refName string) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(3); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(refType)); err != nil {
		return "", fmt.Errorf("sending ref type failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(refName)); err != nil {
		return "", fmt.Errorf("sending ref name failed: %w", err)
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

func (c *Client) ListBranches(repoPath string) ([]string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return nil, fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(4); err != nil {
		return nil, fmt.Errorf("sending command failed: %w", err)
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
	branches := make([]string, 0, count)
	for range count {
		name, err := c.reader.ReadData()
		if err != nil {
			return nil, fmt.Errorf("reading branch name failed: %w", err)
		}
		branches = append(branches, string(name))
	}
	return branches, nil
}

func (c *Client) FormatPatch(repoPath, commitHex string) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(5); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(commitHex)); err != nil {
		return "", fmt.Errorf("sending commit failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return "", fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return "", Perror(status)
	}
	buf, err := c.reader.ReadData()
	if err != nil {
		return "", fmt.Errorf("reading patch failed: %w", err)
	}
	return string(buf), nil
}

func (c *Client) MergeBase(repoPath, hexA, hexB string) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(7); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(hexA)); err != nil {
		return "", fmt.Errorf("sending oid A failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(hexB)); err != nil {
		return "", fmt.Errorf("sending oid B failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return "", fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return "", Perror(status)
	}
	base, err := c.reader.ReadData()
	if err != nil {
		return "", fmt.Errorf("reading base oid failed: %w", err)
	}
	return hex.EncodeToString(base), nil
}

func (c *Client) Log(repoPath, refSpec string, n uint) ([]Commit, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return nil, fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(8); err != nil {
		return nil, fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(refSpec)); err != nil {
		return nil, fmt.Errorf("sending refspec failed: %w", err)
	}
	if err := c.writer.WriteUint(uint64(n)); err != nil {
		return nil, fmt.Errorf("sending limit failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return nil, fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return nil, Perror(status)
	}
	var out []Commit
	for {
		id, err := c.reader.ReadData()
		if err != nil {
			break
		}
		title, _ := c.reader.ReadData()
		authorName, _ := c.reader.ReadData()
		authorEmail, _ := c.reader.ReadData()
		date, _ := c.reader.ReadData()
		out = append(out, Commit{
			Hash:    hex.EncodeToString(id),
			Author:  string(authorName),
			Email:   string(authorEmail),
			Date:    string(date),
			Message: string(title),
		})
	}
	return out, nil
}

func (c *Client) CommitTreeOID(repoPath, commitHex string) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(12); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(commitHex)); err != nil {
		return "", fmt.Errorf("sending oid failed: %w", err)
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
		return "", fmt.Errorf("reading tree oid failed: %w", err)
	}
	return hex.EncodeToString(id), nil
}

func (c *Client) CommitCreate(repoPath, treeHex string, parents []string, authorName, authorEmail string, when time.Time, message string) (string, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return "", fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(13); err != nil {
		return "", fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(treeHex)); err != nil {
		return "", fmt.Errorf("sending tree oid failed: %w", err)
	}
	if err := c.writer.WriteUint(uint64(len(parents))); err != nil {
		return "", fmt.Errorf("sending parents count failed: %w", err)
	}
	for _, p := range parents {
		if err := c.writer.WriteData([]byte(p)); err != nil {
			return "", fmt.Errorf("sending parent oid failed: %w", err)
		}
	}
	if err := c.writer.WriteData([]byte(authorName)); err != nil {
		return "", fmt.Errorf("sending author name failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(authorEmail)); err != nil {
		return "", fmt.Errorf("sending author email failed: %w", err)
	}
	if err := c.writer.WriteInt(when.Unix()); err != nil {
		return "", fmt.Errorf("sending when failed: %w", err)
	}
	_, offset := when.Zone()
	if err := c.writer.WriteInt(int64(offset / 60)); err != nil {
		return "", fmt.Errorf("sending tz offset failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(message)); err != nil {
		return "", fmt.Errorf("sending message failed: %w", err)
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
		return "", fmt.Errorf("reading commit oid failed: %w", err)
	}
	return hex.EncodeToString(id), nil
}

func (c *Client) UpdateRef(repoPath, refName, commitHex string) error {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(14); err != nil {
		return fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(refName)); err != nil {
		return fmt.Errorf("sending ref name failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(commitHex)); err != nil {
		return fmt.Errorf("sending commit oid failed: %w", err)
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

func (c *Client) CommitInfo(repoPath, commitHex string) (*CommitInfo, error) {
	if err := c.writer.WriteData([]byte(repoPath)); err != nil {
		return nil, fmt.Errorf("sending repo path failed: %w", err)
	}
	if err := c.writer.WriteUint(6); err != nil {
		return nil, fmt.Errorf("sending command failed: %w", err)
	}
	if err := c.writer.WriteData([]byte(commitHex)); err != nil {
		return nil, fmt.Errorf("sending commit failed: %w", err)
	}
	status, err := c.reader.ReadUint()
	if err != nil {
		return nil, fmt.Errorf("reading status failed: %w", err)
	}
	if status != 0 {
		return nil, Perror(status)
	}
	id, err := c.reader.ReadData()
	if err != nil {
		return nil, fmt.Errorf("reading id failed: %w", err)
	}
	aname, err := c.reader.ReadData()
	if err != nil {
		return nil, fmt.Errorf("reading author name failed: %w", err)
	}
	aemail, err := c.reader.ReadData()
	if err != nil {
		return nil, fmt.Errorf("reading author email failed: %w", err)
	}
	awhen, err := c.reader.ReadI64()
	if err != nil {
		return nil, fmt.Errorf("reading author time failed: %w", err)
	}
	aoff, err := c.reader.ReadI64()
	if err != nil {
		return nil, fmt.Errorf("reading author tz failed: %w", err)
	}
	cname, err := c.reader.ReadData()
	if err != nil {
		return nil, fmt.Errorf("reading committer name failed: %w", err)
	}
	cemail, err := c.reader.ReadData()
	if err != nil {
		return nil, fmt.Errorf("reading committer email failed: %w", err)
	}
	cwhen, err := c.reader.ReadI64()
	if err != nil {
		return nil, fmt.Errorf("reading committer time failed: %w", err)
	}
	coff, err := c.reader.ReadI64()
	if err != nil {
		return nil, fmt.Errorf("reading committer tz failed: %w", err)
	}
	msg, err := c.reader.ReadData()
	if err != nil {
		return nil, fmt.Errorf("reading message failed: %w", err)
	}
	pcnt, err := c.reader.ReadUint()
	if err != nil {
		return nil, fmt.Errorf("reading parents count failed: %w", err)
	}
	parents := make([]string, 0, pcnt)
	for i := uint64(0); i < pcnt; i++ {
		praw, perr := c.reader.ReadData()
		if perr != nil {
			return nil, fmt.Errorf("reading parent failed: %w", perr)
		}
		parents = append(parents, hex.EncodeToString(praw))
	}
	fcnt, err := c.reader.ReadUint()
	if err != nil {
		return nil, fmt.Errorf("reading file count failed: %w", err)
	}
	files := make([]FileDiff, 0, fcnt)
	for i := uint64(0); i < fcnt; i++ {
		fromMode, err := c.reader.ReadUint()
		if err != nil {
			return nil, fmt.Errorf("reading from mode failed: %w", err)
		}
		toMode, err := c.reader.ReadUint()
		if err != nil {
			return nil, fmt.Errorf("reading to mode failed: %w", err)
		}
		fromPath, err := c.reader.ReadData()
		if err != nil {
			return nil, fmt.Errorf("reading from path failed: %w", err)
		}
		toPath, err := c.reader.ReadData()
		if err != nil {
			return nil, fmt.Errorf("reading to path failed: %w", err)
		}
		ccnt, err := c.reader.ReadUint()
		if err != nil {
			return nil, fmt.Errorf("reading chunk count failed: %w", err)
		}
		chunks := make([]DiffChunk, 0, ccnt)
		for j := uint64(0); j < ccnt; j++ {
			op, err := c.reader.ReadUint()
			if err != nil {
				return nil, fmt.Errorf("reading chunk op failed: %w", err)
			}
			content, err := c.reader.ReadData()
			if err != nil {
				return nil, fmt.Errorf("reading chunk content failed: %w", err)
			}
			chunks = append(chunks, DiffChunk{Op: op, Content: string(content)})
		}
		files = append(files, FileDiff{
			FromMode: fromMode,
			ToMode:   toMode,
			FromPath: string(fromPath),
			ToPath:   string(toPath),
			Chunks:   chunks,
		})
	}
	return &CommitInfo{
		Hash:           hex.EncodeToString(id),
		AuthorName:     string(aname),
		AuthorEmail:    string(aemail),
		AuthorWhen:     awhen,
		AuthorTZMin:    aoff,
		CommitterName:  string(cname),
		CommitterEmail: string(cemail),
		CommitterWhen:  cwhen,
		CommitterTZMin: coff,
		Message:        string(msg),
		Parents:        parents,
		Files:          files,
	}, nil
}
