// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	gliderSSH "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/internal/ansiec"
	"go.lindenii.runxiyu.org/forge/internal/misc"
	goSSH "golang.org/x/crypto/ssh"
)

// serveSSH serves SSH on a [net.Listener]. The listener should generally be a
// TCP listener, although AF_UNIX SOCK_STREAM listeners may be appropriate in
// rare cases.
func (s *Server) serveSSH(listener net.Listener) error {
	var hostKeyBytes []byte
	var hostKey goSSH.Signer
	var err error
	var server *gliderSSH.Server

	if hostKeyBytes, err = os.ReadFile(s.config.SSH.Key); err != nil {
		return err
	}

	if hostKey, err = goSSH.ParsePrivateKey(hostKeyBytes); err != nil {
		return err
	}

	s.serverPubkey = hostKey.PublicKey()
	s.serverPubkeyString = misc.BytesToString(goSSH.MarshalAuthorizedKey(s.serverPubkey))
	s.serverPubkeyFP = goSSH.FingerprintSHA256(s.serverPubkey)

	server = &gliderSSH.Server{
		Handler: func(session gliderSSH.Session) {
			clientPubkey := session.PublicKey()
			var clientPubkeyStr string
			if clientPubkey != nil {
				clientPubkeyStr = strings.TrimSuffix(misc.BytesToString(goSSH.MarshalAuthorizedKey(clientPubkey)), "\n")
			}

			slog.Info("incoming ssh", "addr", session.RemoteAddr().String(), "key", clientPubkeyStr, "command", session.RawCommand())
			fmt.Fprintln(session.Stderr(), ansiec.Blue+"Lindenii Forge "+version+", source at "+strings.TrimSuffix(s.config.HTTP.Root, "/")+"/-/source/"+ansiec.Reset+"\r")

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
				err = s.sshHandleUploadPack(session, clientPubkeyStr, cmd[1])
			case "git-receive-pack":
				if len(cmd) > 2 {
					fmt.Fprintln(session.Stderr(), "Too many arguments\r")
					return
				}
				err = s.sshHandleRecvPack(session, clientPubkeyStr, cmd[1])
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
		slog.Error("error serving SSH", "error", err.Error())
		os.Exit(1)
	}

	return nil
}
