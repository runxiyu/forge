// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"os"
	"os/exec"

	glider_ssh "github.com/gliderlabs/ssh"
)

// sshHandleUploadPack handles clones/fetches. It just uses git-upload-pack
// and has no ACL checks.
func sshHandleUploadPack(session glider_ssh.Session, pubkey, repoIdentifier string) (err error) {
	var repoPath string
	if _, _, _, repoPath, _, _, _, _, err = getRepoInfo2(session.Context(), repoIdentifier, pubkey); err != nil {
		return err
	}

	proc := exec.CommandContext(session.Context(), "git-upload-pack", repoPath)
	proc.Env = append(os.Environ(), "LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket)
	proc.Stdin = session
	proc.Stdout = session
	proc.Stderr = session.Stderr()

	if err = proc.Start(); err != nil {
		fmt.Fprintln(session.Stderr(), "Error while starting process:", err)
		return err
	}

	err = proc.Wait()
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
	}

	return err
}
