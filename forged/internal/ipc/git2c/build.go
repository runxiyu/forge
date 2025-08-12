// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

import (
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strings"
)

func (c *Client) BuildTreeRecursive(repoPath, baseTreeHex string, updates map[string]string) (string, error) {
	treeCache := make(map[string][]TreeEntryRaw)
	var walk func(prefix, hexid string) error
	walk = func(prefix, hexid string) error {
		ents, err := c.TreeListByOID(repoPath, hexid)
		if err != nil {
			return err
		}
		treeCache[prefix] = ents
		for _, e := range ents {
			if e.Mode == 40000 {
				sub := path.Join(prefix, e.Name)
				if err := walk(sub, e.OID); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk("", baseTreeHex); err != nil {
		return "", err
	}

	for p, blob := range updates {
		parts := strings.Split(p, "/")
		dir := strings.Join(parts[:len(parts)-1], "/")
		name := parts[len(parts)-1]
		entries := treeCache[dir]
		found := false
		for i := range entries {
			if entries[i].Name == name {
				if blob == "" {
					entries = append(entries[:i], entries[i+1:]...)
				} else {
					entries[i].Mode = 0o100644
					entries[i].OID = blob
				}
				found = true
				break
			}
		}
		if !found && blob != "" {
			entries = append(entries, TreeEntryRaw{Mode: 0o100644, Name: name, OID: blob})
		}
		treeCache[dir] = entries
	}

	built := make(map[string]string)
	var build func(prefix string) (string, error)
	build = func(prefix string) (string, error) {
		entries := treeCache[prefix]
		for i := range entries {
			if entries[i].Mode == 0o40000 || entries[i].Mode == 40000 {
				sub := path.Join(prefix, entries[i].Name)
				var ok bool
				var oid string
				if oid, ok = built[sub]; !ok {
					var err error
					oid, err = build(sub)
					if err != nil {
						return "", err
					}
				}
				entries[i].Mode = 0o40000
				entries[i].OID = oid
			}
		}
		sort.Slice(entries, func(i, j int) bool {
			ni, nj := entries[i].Name, entries[j].Name
			if ni == nj {
				return entries[i].Mode != 0o40000 && entries[j].Mode == 0o40000
			}
			if strings.HasPrefix(nj, ni) && len(ni) < len(nj) {
				return entries[i].Mode != 0o40000
			}
			if strings.HasPrefix(ni, nj) && len(nj) < len(ni) {
				return entries[j].Mode == 0o40000
			}
			return ni < nj
		})
		wr := make([]TreeEntryRaw, 0, len(entries))
		for _, e := range entries {
			if e.OID == "" {
				continue
			}
			if e.Mode == 40000 {
				e.Mode = 0o40000
			}
			if _, err := hex.DecodeString(e.OID); err != nil {
				return "", fmt.Errorf("invalid OID hex for %s/%s: %w", prefix, e.Name, err)
			}
			wr = append(wr, TreeEntryRaw{Mode: e.Mode, Name: e.Name, OID: e.OID})
		}
		id, err := c.WriteTree(repoPath, wr)
		if err != nil {
			return "", err
		}
		built[prefix] = id
		return id, nil
	}
	root, err := build("")
	if err != nil {
		return "", err
	}
	return root, nil
}
