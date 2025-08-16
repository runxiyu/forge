package web

import (
	"fmt"
	"net/http"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Server struct {
	net        string
	addr       string
	root       string
	httpServer *http.Server
}

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

type Config struct {
	Net            string `scfg:"net"`
	Addr           string `scfg:"addr"`
	Root           string `scfg:"root"`
	CookieExpiry   int    `scfg:"cookie_expiry"`
	ReadTimeout    uint32 `scfg:"read_timeout"`
	WriteTimeout   uint32 `scfg:"write_timeout"`
	IdleTimeout    uint32 `scfg:"idle_timeout"`
	MaxHeaderBytes int    `scfg:"max_header_bytes"`
	ReverseProxy   bool   `scfg:"reverse_proxy"`
}

func New(config Config) (server *Server) {
	handler := &handler{}
	return &Server{
		net:  config.Net,
		addr: config.Addr,
		root: config.Root,
		httpServer: &http.Server{
			Handler:        handler,
			ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(config.WriteTimeout) * time.Second,
			IdleTimeout:    time.Duration(config.IdleTimeout) * time.Second,
			MaxHeaderBytes: config.MaxHeaderBytes,
		},
	}
}

func (server *Server) Run() (err error) {
	listener, err := misc.Listen(server.net, server.addr)
	if err = server.httpServer.Serve(listener); err != nil {
		return fmt.Errorf("serve web: %w", err)
	}
	panic("unreachable")
}
