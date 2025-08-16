// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package git2c

// Commit represents a single commit object retrieved from the git2d daemon.
type Commit struct {
	Hash    string
	Author  string
	Email   string
	Date    string
	Message string
}

// FilenameContents holds the filename and byte contents of a file, such as a README.
type FilenameContents struct {
	Filename string
	Content  []byte
}

// TreeEntry represents a file or directory entry within a Git tree object.
type TreeEntry struct {
	Name      string
	Mode      string
	Size      uint64
	IsFile    bool
	IsSubtree bool
}
