// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gitcmd

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// AdvertiseUploadPack streams advertised references for a repository.
func AdvertiseUploadPack(ctx context.Context, repoPath string, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "git", "upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	cmd.Stderr = cmd.Stdout
	if err = cmd.Start(); err != nil {
		return err
	}
	if err = packLine(w, "# service=git-upload-pack\n"); err != nil {
		return err
	}
	if err = packFlush(w); err != nil {
		return err
	}
	if _, err = io.Copy(w, stdout); err != nil {
		return err
	}
	return cmd.Wait()
}

// Taken from https://github.com/icyphox/legit, MIT license.
func packLine(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "%04x%s", len(s)+4, s)
	return err
}

// Taken from https://github.com/icyphox/legit, MIT license.
func packFlush(w io.Writer) error {
	_, err := fmt.Fprint(w, "0000")
	return err
}
