// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"io"
	"io/fs"
	"os"
)

func (s *server) deployGit2D() (err error) {
	var srcFD fs.File
	var dstFD *os.File

	if srcFD, err = resourcesFS.Open("git2d/git2d"); err != nil {
		return err
	}
	defer srcFD.Close()

	if dstFD, err = os.OpenFile(s.config.Git.DaemonPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755); err != nil {
		return err
	}
	defer dstFD.Close()

	_, err = io.Copy(dstFD, srcFD)

	return err
}
