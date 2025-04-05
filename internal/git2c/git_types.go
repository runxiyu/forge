// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

type Commit struct {
	Hash    string
	Author  string
	Email   string
	Date    string
	Message string
}

type FilenameContents struct {
	Filename string
	Content  []byte
}

type TreeEntry struct {
	Name      string
	Mode      string
	Size      uint64
	IsFile    bool
	IsSubtree bool
}
