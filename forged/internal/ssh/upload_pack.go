// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package ssh

import (
	"fmt"

	glider_ssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/gitcmd"
)

// sshHandleUploadPack handles clones/fetches. It just uses git-upload-pack
// and has no ACL checks.
func (s *Server) sshHandleUploadPack(session glider_ssh.Session, pubkey, repoIdentifier string) (err error) {
	var repoPath string
	if _, _, _, repoPath, _, _, _, _, err = s.getRepoInfo2(session.Context(), repoIdentifier, pubkey); err != nil {
		return err
	}

	err = gitcmd.UploadPack(session.Context(), repoPath,
		[]string{"LINDENII_FORGE_HOOKS_SOCKET_PATH=" + s.config.Hooks.Socket},
		session, session, session.Stderr())
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
	}

	return err
}
