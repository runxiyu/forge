package hooks

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/cmap"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Server struct {
	hookMap         cmap.Map[string, hookInfo]
	socketPath      string
	executablesPath string
}

type Config struct {
	Socket string `scfg:"socket"`
	Execs  string `scfg:"execs"`
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

func New(config Config) (server *Server) {
	return &Server{
		socketPath:      config.Socket,
		executablesPath: config.Execs,
	}
}

func (server *Server) Run(ctx context.Context) error {
	listener, _, err := misc.ListenUnixSocket(server.socketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket for hooks: %w", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
		// TODO: Log the error
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept conn: %w", err)
		}

		go server.handleConn(conn)
	}
}

func (server *Server) handleConn(conn net.Conn) {
	panic("TODO: handle hook connection")
}
