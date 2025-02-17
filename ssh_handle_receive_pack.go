package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"os/exec"

	glider_ssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
)

var err_unauthorized_push = errors.New("You are not authorized to push to this repository")

type pack_to_hook_t struct {
	session       *glider_ssh.Session
	pubkey        string
	direct_access bool
	repo_path     string
}

var pack_to_hook_by_cookie = cmap.Map[string, pack_to_hook_t]{}

func ssh_handle_receive_pack(session glider_ssh.Session, pubkey string, repo_identifier string) (err error) {
	repo_path, access, err := get_repo_path_perms_from_ssh_path_pubkey(session.Context(), repo_identifier, pubkey)
	if err != nil {
		return err
	}

	cookie, err := random_urlsafe_string(16)
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while generating cookie:", err)
	}

	pack_to_hook_by_cookie.Store(cookie, pack_to_hook_t{
		session:       &session,
		pubkey:        pubkey,
		direct_access: access,
		repo_path:     repo_path,
	})
	defer pack_to_hook_by_cookie.Delete(cookie)

	proc := exec.CommandContext(session.Context(), "git-receive-pack", repo_path)
	proc.Env = append(os.Environ(),
		"LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket,
		"LINDENII_FORGE_HOOKS_COOKIE="+cookie,
	)
	proc.Stdin = session
	proc.Stdout = session
	proc.Stderr = session.Stderr()

	err = proc.Start()
	if err != nil {
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

func random_string(sz int) (string, error) {
	r := make([]byte, sz)
	_, err := rand.Read(r)
	return string(r), err
}
