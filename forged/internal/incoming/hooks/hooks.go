package hooks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/cmap"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/global"
)

type Server struct {
	hookMap         cmap.Map[string, hookInfo]
	socketPath      string
	executablesPath string
	global          *global.Global
}
type hookInfo struct {
	session      ssh.Session
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

func New(global *global.Global) (server *Server) {
	cfg := global.Config.Hooks
	return &Server{
		socketPath:      cfg.Socket,
		executablesPath: cfg.Execs,
		hookMap:         cmap.Map[string, hookInfo]{},
		global:          global,
	}
}

func (server *Server) Run(ctx context.Context) error {
	listener, _, err := misc.ListenUnixSocket(ctx, server.socketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket for hooks: %w", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	stop := context.AfterFunc(ctx, func() {
		_ = listener.Close()
	})
	defer stop()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept conn: %w", err)
		}

		go server.handleConn(ctx, conn)
	}
}

func (server *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	unblock := context.AfterFunc(ctx, func() {
		_ = conn.SetDeadline(time.Now())
		_ = conn.Close()
	})
	defer unblock()
}
