package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/global"
)

type Server struct {
	net             string
	addr            string
	root            string
	httpServer      *http.Server
	shutdownTimeout uint32
	global          *global.Global
}

func New(config Config, global *global.Global, queries *queries.Queries) *Server {
	httpServer := &http.Server{
		Handler:        NewHandler(config, global, queries),
		ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(config.IdleTimeout) * time.Second,
		MaxHeaderBytes: config.MaxHeaderBytes,
	} //exhaustruct:ignore
	return &Server{
		net:             config.Net,
		addr:            config.Addr,
		root:            config.Root,
		shutdownTimeout: config.ShutdownTimeout,
		httpServer:      httpServer,
		global:          global,
	}
}

func (server *Server) Run(ctx context.Context) (err error) {
	server.httpServer.BaseContext = func(_ net.Listener) context.Context { return ctx }

	listener, err := misc.Listen(ctx, server.net, server.addr)
	if err != nil {
		return fmt.Errorf("listen for web: %w", err)
	}
	defer func() {
		_ = listener.Close()
	}()

	stop := context.AfterFunc(ctx, func() {
		shCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Duration(server.shutdownTimeout)*time.Second)
		defer cancel()
		_ = server.httpServer.Shutdown(shCtx)
		_ = listener.Close()
	})
	defer stop()

	err = server.httpServer.Serve(listener)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) || ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("serve web: %w", err)
	}
	panic("unreachable")
}
