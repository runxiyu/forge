package config

import (
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/irc"
)

type Config struct {
	HTTP struct {
		Net          string `scfg:"net"`
		Addr         string `scfg:"addr"`
		CookieExpiry int    `scfg:"cookie_expiry"`
		Root         string `scfg:"root"`
		ReadTimeout  uint32 `scfg:"read_timeout"`
		WriteTimeout uint32 `scfg:"write_timeout"`
		IdleTimeout  uint32 `scfg:"idle_timeout"`
		ReverseProxy bool   `scfg:"reverse_proxy"`
	} `scfg:"http"`
	Hooks struct {
		Socket string `scfg:"socket"`
		Execs  string `scfg:"execs"`
	} `scfg:"hooks"`
	LMTP struct {
		Socket       string `scfg:"socket"`
		Domain       string `scfg:"domain"`
		MaxSize      int64  `scfg:"max_size"`
		WriteTimeout uint32 `scfg:"write_timeout"`
		ReadTimeout  uint32 `scfg:"read_timeout"`
	} `scfg:"lmtp"`
	Git struct {
		RepoDir    string `scfg:"repo_dir"`
		Socket     string `scfg:"socket"`
		DaemonPath string `scfg:"daemon_path"`
	} `scfg:"git"`
	SSH struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
		Key  string `scfg:"key"`
		Root string `scfg:"root"`
	} `scfg:"ssh"`
	IRC     irc.Config `scfg:"irc"`
	General struct {
		Title string `scfg:"title"`
	} `scfg:"general"`
	DB database.Config `scfg:"db"`
	Pprof struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
	} `scfg:"pprof"`
}
