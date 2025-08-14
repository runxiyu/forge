// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package ssh

import (
	"errors"
	"fmt"

	gliderSSH "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/gitcmd"
	"go.lindenii.runxiyu.org/forge/forged/internal/gogit"
	"go.lindenii.runxiyu.org/forge/forged/internal/models"
)

// packPass contains information known when handling incoming SSH connections
// that then needs to be used in hook socket connection handlers. See hookc(1).
type packPass = PackPass

// sshHandleRecvPack handles attempts to push to repos.
func (s *Server) sshHandleRecvPack(session gliderSSH.Session, pubkey, repoIdentifier string) (err error) {
	groupPath, repoName, repoID, repoPath, directAccess, contribReq, userType, userID, err := s.getRepoInfo2(session.Context(), repoIdentifier, pubkey)
	if err != nil {
		return err
	}
	repo, err := gogit.Open(repoPath)
	if err != nil {
		return err
	}

	repoConf, err := repo.Config()
	if err != nil {
		return err
	}

	repoConfCore := repoConf.Raw.Section("core")
	if repoConfCore == nil {
		return errors.New("repository has no core section in config")
	}

	hooksPath := repoConfCore.OptionAll("hooksPath")
	if len(hooksPath) != 1 || hooksPath[0] != s.config.Hooks.Execs {
		return errors.New("repository has hooksPath set to an unexpected value")
	}

	if !directAccess {
		switch contribReq {
		case "closed":
			if !directAccess {
				return errors.New("you need direct access to push to this repo")
			}
		case "registered_user":
			if userType != "registered" {
				return errors.New("you need to be a registered user to push to this repo")
			}
		case "ssh_pubkey":
			fallthrough
		case "federated":
			if pubkey == "" {
				return errors.New("you need to have an SSH public key to push to this repo")
			}
			if userType == "" {
				userID, err = models.AddUserSSH(session.Context(), s.database, pubkey)
				if err != nil {
					return err
				}
				fmt.Fprintln(session.Stderr(), "you are now registered as user ID", userID)
				userType = "pubkey_only"
			}

		case "public":
		default:
			panic("unknown contrib_requirements value " + contribReq)
		}
	}

	cookie, err := randomUrlsafeStr(16)
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while generating cookie:", err)
	}

	s.packPasses.Store(cookie, packPass{
		Session:      session,
		Pubkey:       pubkey,
		DirectAccess: directAccess,
		RepoPath:     repoPath,
		UserID:       userID,
		RepoID:       repoID,
		GroupPath:    groupPath,
		RepoName:     repoName,
		Repo:         repo,
		ContribReq:   contribReq,
		UserType:     userType,
	})
	defer s.packPasses.Delete(cookie)
	// The Delete won't execute until proc.Wait returns unless something
	// horribly wrong such as a panic occurs.

	err = gitcmd.ReceivePack(session.Context(), repoPath,
		[]string{
			"LINDENII_FORGE_HOOKS_SOCKET_PATH=" + s.config.Hooks.Socket,
			"LINDENII_FORGE_HOOKS_COOKIE=" + cookie,
		}, session, session, session.Stderr())
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
	}

	return err
}
