package lmtp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Server struct {
	socket       string
	domain       string
	maxSize      int64
	writeTimeout uint32
	readTimeout  uint32
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

func (server *Server) Run(ctx context.Context) error {
	listener, _, err := misc.ListenUnixSocket(ctx, server.socket)
	if err != nil {
		return fmt.Errorf("listen unix socket for LMTP: %w", err)
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
