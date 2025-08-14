// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gitcmd

import (
	"log"
	"os/exec"
)

// RunDaemon runs the Git2D daemon at the given path and socket.
func RunDaemon(path, socket string) error {
	cmd := exec.Command(path, socket) //#nosec G204
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}
