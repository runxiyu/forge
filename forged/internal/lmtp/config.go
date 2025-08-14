package lmtp

import (
	"fmt"
	"net"

	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

type Pool struct {
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

func New(config Config) (pool Pool) {
	pool.socket = config.Socket
	pool.domain = config.Domain
	pool.maxSize = config.MaxSize
	pool.writeTimeout = config.WriteTimeout
	pool.readTimeout = config.ReadTimeout
	return pool
}

func (pool *Pool) Run() error {
	listener, _, err := misc.ListenUnixSocket(pool.socket)
	if err != nil {
		return fmt.Errorf("listen unix socket for LMTP: %w", err)
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
	panic("TODO: handle LMTP connection")
}
