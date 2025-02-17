package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	glider_ssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
)

var err_unauthorized_push = errors.New("You are not authorized to push to this repository")

type hooks_cookie_deployer_return struct {
	args     []string
	callback chan struct{}
	conn     net.Conn
}

var hooks_cookie_deployer = cmap.ComparableMap[string, chan hooks_cookie_deployer_return]{}

func ssh_handle_receive_pack(session glider_ssh.Session, pubkey string, repo_identifier string) (err error) {
	repo_path, access, err := get_repo_path_perms_from_ssh_path_pubkey(session.Context(), repo_identifier, pubkey)
	if err != nil {
		return err
	}
	if !access {
		return err_unauthorized_push
	}

	cookie, err := random_urlsafe_string(16)
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while generating cookie:", err)
	}

	deployer_channel := make(chan hooks_cookie_deployer_return)
	hooks_cookie_deployer.Store(cookie, deployer_channel)
	defer hooks_cookie_deployer.Delete(cookie)

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

	deployer := <-deployer_channel

	deployer.conn.Write([]byte{0})

	deployer.callback <- struct{}{}

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
