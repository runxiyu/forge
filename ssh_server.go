// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	gliderSSH "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/lindenii-common/ansiec"
	"go.lindenii.runxiyu.org/lindenii-common/clog"
	goSSH "golang.org/x/crypto/ssh"
)

var (
	serverPubkeyString string
	serverPubkeyFP     string
	serverPubkey       goSSH.PublicKey
)

func serveSSH(listener net.Listener) error {
	var host_key_bytes []byte
	var host_key goSSH.Signer
	var err error
	var server *gliderSSH.Server

	if host_key_bytes, err = os.ReadFile(config.SSH.Key); err != nil {
		return err
	}

	if host_key, err = goSSH.ParsePrivateKey(host_key_bytes); err != nil {
		return err
	}

	serverPubkey = host_key.PublicKey()
	serverPubkeyString = string(goSSH.MarshalAuthorizedKey(serverPubkey))
	serverPubkeyFP = goSSH.FingerprintSHA256(serverPubkey)

	server = &gliderSSH.Server{
		Handler: func(session gliderSSH.Session) {
			client_public_key := session.PublicKey()
			var client_public_key_string string
			if client_public_key != nil {
				client_public_key_string = strings.TrimSuffix(string(goSSH.MarshalAuthorizedKey(client_public_key)), "\n")
			}

			clog.Info("Incoming SSH: " + session.RemoteAddr().String() + " " + client_public_key_string + " " + session.RawCommand())
			fmt.Fprintln(session.Stderr(), ansiec.Blue+"Lindenii Forge "+VERSION+", source at "+strings.TrimSuffix(config.HTTP.Root, "/")+"/:/source/"+ansiec.Reset+"\r")

			cmd := session.Command()

			if len(cmd) < 2 {
				fmt.Fprintln(session.Stderr(), "Insufficient arguments\r")
				return
			}

			switch cmd[0] {
			case "git-upload-pack":
				if len(cmd) > 2 {
					fmt.Fprintln(session.Stderr(), "Too many arguments\r")
					return
				}
				err = sshHandleUploadPack(session, client_public_key_string, cmd[1])
			case "git-receive-pack":
				if len(cmd) > 2 {
					fmt.Fprintln(session.Stderr(), "Too many arguments\r")
					return
				}
				err = sshHandleRecvPack(session, client_public_key_string, cmd[1])
			default:
				fmt.Fprintln(session.Stderr(), "Unsupported command: "+cmd[0]+"\r")
				return
			}
			if err != nil {
				fmt.Fprintln(session.Stderr(), err.Error())
				return
			}
		},
		PublicKeyHandler:           func(ctx gliderSSH.Context, key gliderSSH.PublicKey) bool { return true },
		KeyboardInteractiveHandler: func(ctx gliderSSH.Context, challenge goSSH.KeyboardInteractiveChallenge) bool { return true },
		// It is intentional that we do not check any credentials and accept all connections.
		// This allows all users to connect and clone repositories. However, the public key
		// is passed to handlers, so e.g. the push handler could check the key and reject the
		// push if it needs to.
	}

	server.AddHostKey(host_key)

	if err = server.Serve(listener); err != nil {
		clog.Fatal(1, "Serving SSH: "+err.Error())
	}

	return nil
}
