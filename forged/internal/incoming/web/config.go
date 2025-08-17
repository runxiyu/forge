package web

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
	TemplatesPath   string `scfg:"templates_path"`
	StaticPath      string `scfg:"static_path"`
}
