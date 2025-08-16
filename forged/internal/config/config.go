package config

import (
	"bufio"
	"log/slog"
	"os"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/scfg"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/hooks"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/lmtp"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/ssh"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/irc"
)

type Config struct {
	DB    database.Config `scfg:"db"`
	Web   web.Config      `scfg:"web"`
	Hooks hooks.Config    `scfg:"hooks"`
	LMTP  lmtp.Config     `scfg:"lmtp"`
	SSH   ssh.Config      `scfg:"ssh"`
	IRC   irc.Config      `scfg:"irc"`
	Git   struct {
		RepoDir    string `scfg:"repo_dir"`
		Socket     string `scfg:"socket"`
		DaemonPath string `scfg:"daemon_path"`
	} `scfg:"git"`
	General struct {
		Title string `scfg:"title"`
	} `scfg:"general"`
	Pprof struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
	} `scfg:"pprof"`
}

func Open(path string) (config Config, err error) {
	var configFile *os.File

	if configFile, err = os.Open(path); err != nil {
		return config, err
	}
	defer configFile.Close()

	decoder := scfg.NewDecoder(bufio.NewReader(configFile))
	if err = decoder.Decode(&config); err != nil {
		return config, err
	}
	for _, u := range decoder.UnknownDirectives() {
		slog.Warn("unknown configuration directive", "directive", u)
	}

	return config, err
}
