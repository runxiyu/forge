package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	glider_ssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
	go_ssh "golang.org/x/crypto/ssh"
)

var (
	server_public_key_string      string
	server_public_key_fingerprint string
	server_public_key             go_ssh.PublicKey
)

func serve_ssh(listener net.Listener) error {
	host_key_bytes, err := os.ReadFile(config.SSH.Key)
	if err != nil {
		return err
	}

	host_key, err := go_ssh.ParsePrivateKey(host_key_bytes)
	if err != nil {
		return err
	}

	server_public_key = host_key.PublicKey()
	server_public_key_string = string(go_ssh.MarshalAuthorizedKey(server_public_key))
	server_public_key_fingerprint = string(go_ssh.FingerprintSHA256(server_public_key))

	server := &glider_ssh.Server{
		Handler: func(session glider_ssh.Session) {
			client_public_key := session.PublicKey()
			var client_public_key_string string
			if client_public_key != nil {
				client_public_key_string = string(go_ssh.MarshalAuthorizedKey(client_public_key))
			}

			clog.Info("Incoming SSH: " + session.RemoteAddr().String() + " " + strings.TrimSuffix(client_public_key_string, "\n") + " " + session.RawCommand())
			fmt.Fprintln(session.Stderr(), "Lindenii Forge " + VERSION + ", source at " + strings.TrimSuffix(config.HTTP.Root, "/") + "/:/source/")

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
			if exit_error, ok := err.(*exec.ExitError); ok {
				fmt.Fprintln(session.Stderr(), "Process exited with error", exit_error.ExitCode())
			} else if err != nil {
				fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
			}
		},
		PublicKeyHandler:           func(ctx glider_ssh.Context, key glider_ssh.PublicKey) bool { return true },
		KeyboardInteractiveHandler: func(ctx glider_ssh.Context, challenge go_ssh.KeyboardInteractiveChallenge) bool { return true },
		// It is intentional that we do not check any credentials and accept all connections.
		// This allows all users to connect and clone repositories; when pushing is added later,
		// we will check their public key in the session handler, not in the auth handlers.
	}

	server.AddHostKey(host_key)

	go func() {
		err = server.Serve(listener)
		if err != nil {
			clog.Fatal(1, "Serving SSH: "+err.Error())
		}
	}()

	return nil
}
