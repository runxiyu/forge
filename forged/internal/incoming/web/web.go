package web

import "net/http"

type Server struct {
	httpServer *http.Server
}

type Config struct {
	Net          string `scfg:"net"`
	Addr         string `scfg:"addr"`
	CookieExpiry int    `scfg:"cookie_expiry"`
	Root         string `scfg:"root"`
	ReadTimeout  uint32 `scfg:"read_timeout"`
	WriteTimeout uint32 `scfg:"write_timeout"`
	IdleTimeout  uint32 `scfg:"idle_timeout"`
	ReverseProxy bool   `scfg:"reverse_proxy"`
}
