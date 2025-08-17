package config

import (
	"bufio"
	"fmt"
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
