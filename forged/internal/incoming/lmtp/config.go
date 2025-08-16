package lmtp

import (
	"fmt"
	"net"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Server struct {
	socket       string
	domain       string
	maxSize      int64
	writeTimeout uint32
	readTimeout  uint32
}

type Config struct {
	Socket       string `scfg:"socket"`
	Domain       string `scfg:"domain"`
	MaxSize      int64  `scfg:"max_size"`
	WriteTimeout uint32 `scfg:"write_timeout"`
	ReadTimeout  uint32 `scfg:"read_timeout"`
}

func New(config Config) (server *Server) {
	return &Server{
		socket:       config.Socket,
		domain:       config.Domain,
		maxSize:      config.MaxSize,
		writeTimeout: config.WriteTimeout,
		readTimeout:  config.ReadTimeout,
	}
}

func (server *Server) Run() error {
	listener, _, err := misc.ListenUnixSocket(server.socket)
	if err != nil {
		return fmt.Errorf("listen unix socket for LMTP: %w", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept conn: %w", err)
		}

		go server.handleConn(conn)
	}
}

func (server *Server) handleConn(conn net.Conn) {
	panic("TODO: handle LMTP connection")
}
