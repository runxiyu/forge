// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	gliderSSH "github.com/gliderlabs/ssh"
	"github.com/go-git/go-git/v5"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
)

type packPass struct {
	session      gliderSSH.Session
	repo         *git.Repository
	pubkey       string
	directAccess bool
	repoPath     string
	userID       int
	userType     string
	repoID       int
	groupPath    []string
	repoName     string
	contribReq   string
}

var packPasses = cmap.Map[string, packPass]{}

// sshHandleRecvPack handles attempts to push to repos.
func sshHandleRecvPack(session gliderSSH.Session, pubkey, repoIdentifier string) (err error) {
	groupPath, repoName, repoID, repoPath, directAccess, contribReq, userType, userID, err := getRepoInfo2(session.Context(), repoIdentifier, pubkey)
	if err != nil {
		return err
	}
	repo, err := git.PlainOpen(repoPath)
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
	if len(hooksPath) != 1 || hooksPath[0] != config.Hooks.Execs {
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
				userID, err = addUserSSH(session.Context(), pubkey)
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

	packPasses.Store(cookie, packPass{
		session:      session,
		pubkey:       pubkey,
		directAccess: directAccess,
		repoPath:     repoPath,
		userID:       userID,
		repoID:       repoID,
		groupPath:    groupPath,
		repoName:     repoName,
		repo:         repo,
		contribReq:   contribReq,
		userType:     userType,
	})
	defer packPasses.Delete(cookie)
	// The Delete won't execute until proc.Wait returns unless something
	// horribly wrong such as a panic occurs.

	proc := exec.CommandContext(session.Context(), "git-receive-pack", repoPath)
	proc.Env = append(os.Environ(),
		"LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket,
		"LINDENII_FORGE_HOOKS_COOKIE="+cookie,
	)
	proc.Stdin = session
	proc.Stdout = session
	proc.Stderr = session.Stderr()

	if err = proc.Start(); err != nil {
		fmt.Fprintln(session.Stderr(), "Error while starting process:", err)
		return err
	}

	err = proc.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		fmt.Fprintln(session.Stderr(), "Process exited with error", exitError.ExitCode())
	} else if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
	}

	return err
}
