// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gitcmd

import (
	"context"
	"io"
	"os"
	"os/exec"
)

// ReceivePack runs git-receive-pack for the given repository path.
func ReceivePack(ctx context.Context, repoPath string, env []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	cmd := exec.CommandContext(ctx, "git-receive-pack", repoPath)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = errOut
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
