package ssh

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	gossh "golang.org/x/crypto/ssh"
)

type Server struct {
	gliderServer    *gliderssh.Server
	privkey         gossh.Signer
	pubkeyString    string
	pubkeyFP        string
	net             string
	addr            string
	root            string
	shutdownTimeout uint32
}

func New(config Config) (server *Server, err error) {
	server = &Server{
		net:             config.Net,
		addr:            config.Addr,
		root:            config.Root,
		shutdownTimeout: config.ShutdownTimeout,
	} //exhaustruct:ignore

	var privkeyBytes []byte

	privkeyBytes, err = os.ReadFile(config.Key)
	if err != nil {
		return server, fmt.Errorf("read SSH private key: %w", err)
	}

	server.privkey, err = gossh.ParsePrivateKey(privkeyBytes)
	if err != nil {
		return server, fmt.Errorf("parse SSH private key: %w", err)
	}

	server.pubkeyString = misc.BytesToString(gossh.MarshalAuthorizedKey(server.privkey.PublicKey()))
	server.pubkeyFP = gossh.FingerprintSHA256(server.privkey.PublicKey())

	server.gliderServer = &gliderssh.Server{
		Handler:                    handle,
		PublicKeyHandler:           func(ctx gliderssh.Context, key gliderssh.PublicKey) bool { return true },
		KeyboardInteractiveHandler: func(ctx gliderssh.Context, challenge gossh.KeyboardInteractiveChallenge) bool { return true },
	} //exhaustruct:ignore
	server.gliderServer.AddHostKey(server.privkey)

	return server, nil
}

func (server *Server) Run(ctx context.Context) (err error) {
	listener, err := misc.Listen(ctx, server.net, server.addr)
	if err != nil {
		return fmt.Errorf("listen for SSH: %w", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	stop := context.AfterFunc(ctx, func() {
		shCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(server.shutdownTimeout)*time.Second)
		defer cancel()
		_ = server.gliderServer.Shutdown(shCtx)
		_ = listener.Close()
	})
	defer stop()

	err = server.gliderServer.Serve(listener)
	if err != nil {
		if errors.Is(err, gliderssh.ErrServerClosed) || ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("serve SSH: %w", err)
	}
	panic("unreachable")
}

func handle(session gliderssh.Session) {
	panic("SSH server handler not implemented yet")
}
