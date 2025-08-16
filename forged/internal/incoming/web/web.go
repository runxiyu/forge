package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Server struct {
	net             string
	addr            string
	root            string
	httpServer      *http.Server
	shutdownTimeout uint32
}

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

type Config struct {
	Net             string `scfg:"net"`
	Addr            string `scfg:"addr"`
	Root            string `scfg:"root"`
	CookieExpiry    int    `scfg:"cookie_expiry"`
	ReadTimeout     uint32 `scfg:"read_timeout"`
	WriteTimeout    uint32 `scfg:"write_timeout"`
	IdleTimeout     uint32 `scfg:"idle_timeout"`
	MaxHeaderBytes  int    `scfg:"max_header_bytes"`
	ReverseProxy    bool   `scfg:"reverse_proxy"`
	ShutdownTimeout uint32 `scfg:"shutdown_timeout"`
}

func New(config Config) (server *Server) {
	handler := &handler{}
	return &Server{
		net:             config.Net,
		addr:            config.Addr,
		root:            config.Root,
		shutdownTimeout: config.ShutdownTimeout,
		httpServer: &http.Server{
			Handler:        handler,
			ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
			IdleTimeout:    time.Duration(config.IdleTimeout) * time.Second,
			MaxHeaderBytes: config.MaxHeaderBytes,
		},
	}
}

func (server *Server) Run(ctx context.Context) (err error) {
	server.httpServer.BaseContext = func(_ net.Listener) context.Context { return ctx }

	listener, err := misc.Listen(server.net, server.addr)
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

	if err = server.httpServer.Serve(listener); err != nil {
		if err == http.ErrServerClosed || ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("serve web: %w", err)
	}
	panic("unreachable")
}
