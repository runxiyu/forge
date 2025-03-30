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
	var hostKeyBytes []byte
	var hostKey goSSH.Signer
	var err error
	var server *gliderSSH.Server

	if hostKeyBytes, err = os.ReadFile(config.SSH.Key); err != nil {
		return err
	}

	if hostKey, err = goSSH.ParsePrivateKey(hostKeyBytes); err != nil {
		return err
	}

	serverPubkey = hostKey.PublicKey()
	serverPubkeyString = bytesToString(goSSH.MarshalAuthorizedKey(serverPubkey))
	serverPubkeyFP = goSSH.FingerprintSHA256(serverPubkey)

	server = &gliderSSH.Server{
		Handler: func(session gliderSSH.Session) {
			clientPubkey := session.PublicKey()
			var clientPubkeyStr string
			if clientPubkey != nil {
				clientPubkeyStr = strings.TrimSuffix(bytesToString(goSSH.MarshalAuthorizedKey(clientPubkey)), "\n")
			}

			clog.Info("Incoming SSH: " + session.RemoteAddr().String() + " " + clientPubkeyStr + " " + session.RawCommand())
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
				err = sshHandleUploadPack(session, clientPubkeyStr, cmd[1])
			case "git-receive-pack":
				if len(cmd) > 2 {
					fmt.Fprintln(session.Stderr(), "Too many arguments\r")
					return
				}
				err = sshHandleRecvPack(session, clientPubkeyStr, cmd[1])
			default:
				fmt.Fprintln(session.Stderr(), "Unsupported command: "+cmd[0]+"\r")
				return
			}
			if err != nil {
				fmt.Fprintln(session.Stderr(), err.Error())
				return
			}
		},
		PublicKeyHandler:           func(_ gliderSSH.Context, _ gliderSSH.PublicKey) bool { return true },
		KeyboardInteractiveHandler: func(_ gliderSSH.Context, _ goSSH.KeyboardInteractiveChallenge) bool { return true },
		// It is intentional that we do not check any credentials and accept all connections.
		// This allows all users to connect and clone repositories. However, the public key
		// is passed to handlers, so e.g. the push handler could check the key and reject the
		// push if it needs to.
	} //exhaustruct:ignore

	server.AddHostKey(hostKey)

	if err = server.Serve(listener); err != nil {
		clog.Fatal(1, "Serving SSH: "+err.Error())
	}

	return nil
}
