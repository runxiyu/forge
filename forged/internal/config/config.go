package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/scfg"
)

type Config struct {
	DB      DB      `scfg:"db"`
	Web     Web     `scfg:"web"`
	Hooks   Hooks   `scfg:"hooks"`
	LMTP    LMTP    `scfg:"lmtp"`
	SSH     SSH     `scfg:"ssh"`
	IRC     IRC     `scfg:"irc"`
	Git     Git     `scfg:"git"`
	General General `scfg:"general"`
	Pprof   Pprof   `scfg:"pprof"`
}

type DB struct {
	Conn string `scfg:"conn"`
}

type Web struct {
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

type Hooks struct {
	Socket string `scfg:"socket"`
	Execs  string `scfg:"execs"`
}

type LMTP struct {
	Socket       string `scfg:"socket"`
	Domain       string `scfg:"domain"`
	MaxSize      int64  `scfg:"max_size"`
	WriteTimeout uint32 `scfg:"write_timeout"`
	ReadTimeout  uint32 `scfg:"read_timeout"`
}

type SSH struct {
	Net             string `scfg:"net"`
	Addr            string `scfg:"addr"`
	Key             string `scfg:"key"`
	Root            string `scfg:"root"`
	ShutdownTimeout uint32 `scfg:"shutdown_timeout"`
}

type IRC struct {
	Net   string `scfg:"net"`
	Addr  string `scfg:"addr"`
	TLS   bool   `scfg:"tls"`
	SendQ uint   `scfg:"sendq"`
	Nick  string `scfg:"nick"`
	User  string `scfg:"user"`
	Gecos string `scfg:"gecos"`
}

type Git struct {
	RepoDir string `scfg:"repo_dir"`
	Socket  string `scfg:"socket"`
}

type General struct {
	Title string `scfg:"title"`
}

type Pprof struct {
	Net  string `scfg:"net"`
	Addr string `scfg:"addr"`
}

func Open(path string) (config Config, err error) {
	var configFile *os.File

	configFile, err = os.Open(path) //#nosec G304
	if err != nil {
		err = fmt.Errorf("open config file: %w", err)
		return config, err
	}
	defer func() {
		_ = configFile.Close()
	}()

	decoder := scfg.NewDecoder(bufio.NewReader(configFile))
	err = decoder.Decode(&config)
	if err != nil {
		err = fmt.Errorf("decode config file: %w", err)
		return config, err
	}
	for _, u := range decoder.UnknownDirectives() {
		slog.Warn("unknown configuration directive", "directive", u)
	}

	return config, err
}
