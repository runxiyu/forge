// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gitcmd

import (
	"context"
	"io"
	"os"
	"os/exec"
)

// Run executes a git command with the provided arguments.
func Run(ctx context.Context, env []string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
