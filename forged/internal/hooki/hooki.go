package hooki

import (
	"fmt"
	"net"

	"github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/cmap"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

type Pool struct {
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

func New(config Config) (pool Pool) {
	pool.socketPath = config.Socket
	pool.executablesPath = config.Execs
	return
}

func (pool *Pool) Run() error {
	listener, _, err := misc.ListenUnixSocket(pool.socketPath)
	if err != nil {
		return fmt.Errorf("listen unix socket for hooks: %w", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept conn: %w", err)
		}

		go pool.handleConn(conn)
	}
}

func (pool *Pool) handleConn(conn net.Conn) {
	panic("TODO: handle hook connection")
}
