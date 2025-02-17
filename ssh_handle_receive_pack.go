package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	glider_ssh "github.com/gliderlabs/ssh"
)

var err_unauthorized_push = errors.New("You are not authorized to push to this repository")

func ssh_handle_receive_pack(session glider_ssh.Session, pubkey string, repo_identifier string) (err error) {
	repo_path, access, err := get_repo_path_perms_from_ssh_path_pubkey(session.Context(), repo_identifier, pubkey)
	if err != nil {
		return err
	}
	if !access {
		return err_unauthorized_push
	}

	proc := exec.CommandContext(session.Context(), "git-receive-pack", repo_path)
	proc.Env = append(os.Environ(), "LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket)
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
