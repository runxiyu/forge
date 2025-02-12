package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	glider_ssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
	go_ssh "golang.org/x/crypto/ssh"
)

func serve_ssh() error {
	hostKeyBytes, err := os.ReadFile(config.SSH.Key)
	if err != nil {
		return err
	}

	hostKey, err := go_ssh.ParsePrivateKey(hostKeyBytes)
	if err != nil {
		return err
	}

	server := &glider_ssh.Server{
		Handler: func(session glider_ssh.Session) {
			client_public_key := session.PublicKey()
			var client_public_key_string string
			if client_public_key != nil {
				client_public_key_string = string(go_ssh.MarshalAuthorizedKey(client_public_key))
			}
			_ = client_public_key_string

			cmd := session.Command()

			if len(cmd) < 2 {
				fmt.Fprintln(session.Stderr(), "Insufficient arguments")
				return
			}

			if cmd[0] != "git-upload-pack" {
				fmt.Fprintln(session.Stderr(), "Unsupported command")
				return
			}

			fs_path, err := get_repo_path_from_ssh_path(session.Context(), cmd[1])
			if err != nil {
				fmt.Fprintln(session.Stderr(), "Error while getting repo path:", err)
				return
			}

			proc := exec.CommandContext(session.Context(), cmd[0], fs_path)
			proc.Stdin = session
			proc.Stdout = session
			proc.Stderr = session.Stderr()

			err = proc.Start()
			if err != nil {
				fmt.Fprintln(session.Stderr(), "Error while starting process:", err)
				return
			}
			err = proc.Wait()
			if exitError, ok := err.(*exec.ExitError); ok {
				fmt.Fprintln(session.Stderr(), "Process exited with error", exitError.ExitCode())
			} else if err != nil {
				fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
			}
		},
		PublicKeyHandler:           func(ctx glider_ssh.Context, key glider_ssh.PublicKey) bool { return true },
		KeyboardInteractiveHandler: func(ctx glider_ssh.Context, challenge go_ssh.KeyboardInteractiveChallenge) bool { return true },
	}

	server.AddHostKey(hostKey)

	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		return err
	}

	go func() {
		err = server.Serve(listener)
		if err != nil {
			clog.Fatal(1, "Serving SSH: "+err.Error())
		}
	}()

	return nil
}
