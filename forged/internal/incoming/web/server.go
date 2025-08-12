package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
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

func New(global *global.Global) *Server {
	cfg := global.Config.Web
	httpServer := &http.Server{
		Handler:        NewHandler(global),
		ReadTimeout:    time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	} //exhaustruct:ignore
	return &Server{
		net:             cfg.Net,
		addr:            cfg.Addr,
		root:            cfg.Root,
		shutdownTimeout: cfg.ShutdownTimeout,
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
