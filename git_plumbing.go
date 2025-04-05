// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"

	"go.lindenii.runxiyu.org/forge/internal/misc"
)

func writeTree(ctx context.Context, repoPath string, entries []treeEntry) (string, error) {
	var buf bytes.Buffer

	sort.Slice(entries, func(i, j int) bool {
		nameI, nameJ := entries[i].name, entries[j].name

		if nameI == nameJ { // meh
			return !(entries[i].mode == "40000") && (entries[j].mode == "40000")
		}

		if strings.HasPrefix(nameJ, nameI) && len(nameI) < len(nameJ) {
			return !(entries[i].mode == "40000")
		}

		if strings.HasPrefix(nameI, nameJ) && len(nameJ) < len(nameI) {
			return entries[j].mode == "40000"
		}

		return nameI < nameJ
	})

	for _, e := range entries {
		buf.WriteString(e.mode)
		buf.WriteByte(' ')
		buf.WriteString(e.name)
		buf.WriteByte(0)
		buf.Write(e.sha)
	}

	cmd := exec.CommandContext(ctx, "git", "hash-object", "-w", "-t", "tree", "--stdin")
	cmd.Env = append(os.Environ(), "GIT_DIR="+repoPath)
	cmd.Stdin = &buf

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func buildTreeRecursive(ctx context.Context, repoPath, baseTree string, updates map[string][]byte) (string, error) {
	treeCache := make(map[string][]treeEntry)

	var walk func(string, string) error
	walk = func(prefix, sha string) error {
		cmd := exec.CommandContext(ctx, "git", "cat-file", "tree", sha)
		cmd.Env = append(os.Environ(), "GIT_DIR="+repoPath)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return err
		}
		data := out.Bytes()
		i := 0
		var entries []treeEntry
		for i < len(data) {
			modeEnd := bytes.IndexByte(data[i:], ' ')
			if modeEnd < 0 {
				return errors.New("invalid tree format")
			}
			mode := misc.BytesToString(data[i : i+modeEnd])
			i += modeEnd + 1

			nameEnd := bytes.IndexByte(data[i:], 0)
			if nameEnd < 0 {
				return errors.New("missing null after filename")
			}
			name := misc.BytesToString(data[i : i+nameEnd])
			i += nameEnd + 1

			if i+20 > len(data) {
				return errors.New("unexpected EOF in SHA")
			}
			shaBytes := data[i : i+20]
			i += 20

			entries = append(entries, treeEntry{
				mode: mode,
				name: name,
				sha:  shaBytes,
			})

			if mode == "40000" {
				subPrefix := path.Join(prefix, name)
				if err := walk(subPrefix, hex.EncodeToString(shaBytes)); err != nil {
					return err
				}
			}
		}
		treeCache[prefix] = entries
		return nil
	}

	if err := walk("", baseTree); err != nil {
		return "", err
	}

	for filePath, blobSha := range updates {
		parts := strings.Split(filePath, "/")
		dir := strings.Join(parts[:len(parts)-1], "/")
		name := parts[len(parts)-1]

		entries := treeCache[dir]
		found := false
		for i, e := range entries {
			if e.name == name {
				if blobSha == nil {
					// Remove TODO
					entries = append(entries[:i], entries[i+1:]...)
				} else {
					entries[i].sha = blobSha
				}
				found = true
				break
			}
		}
		if !found && blobSha != nil {
			entries = append(entries, treeEntry{
				mode: "100644",
				name: name,
				sha:  blobSha,
			})
		}
		treeCache[dir] = entries
	}

	built := make(map[string][]byte)
	var build func(string) ([]byte, error)
	build = func(prefix string) ([]byte, error) {
		entries := treeCache[prefix]
		for i, e := range entries {
			if e.mode == "40000" {
				subPrefix := path.Join(prefix, e.name)
				if sha, ok := built[subPrefix]; ok {
					entries[i].sha = sha
					continue
				}
				newShaStr, err := build(subPrefix)
				if err != nil {
					return nil, err
				}
				entries[i].sha = newShaStr
			}
		}
		shaStr, err := writeTree(ctx, repoPath, entries)
		if err != nil {
			return nil, err
		}
		shaBytes, err := hex.DecodeString(shaStr)
		if err != nil {
			return nil, err
		}
		built[prefix] = shaBytes
		return shaBytes, nil
	}

	rootShaBytes, err := build("")
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(rootShaBytes), nil
}

type treeEntry struct {
	mode string // like "100644"
	name string // individual name
	sha  []byte
}
