// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gitcmd

import (
	"context"
	"io"
	"os"
	"os/exec"
)

// UploadPack runs git upload-pack for the given repository path.
func UploadPack(ctx context.Context, repoPath string, env []string, in io.Reader, out io.Writer, errOut io.Writer, args ...string) error {
	cmdArgs := append([]string{"upload-pack"}, args...)
	cmdArgs = append(cmdArgs, repoPath)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = errOut
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
