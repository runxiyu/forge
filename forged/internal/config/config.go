// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package config

import (
	"bufio"
	"log/slog"
	"os"

	"go.lindenii.runxiyu.org/forge/forged/internal/irc"
	"go.lindenii.runxiyu.org/forge/forged/internal/scfg"
)

// Config holds runtime configuration for the Forge server.
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
		RepoDir string `scfg:"repo_dir"`
		Socket  string `scfg:"socket"`
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
	Resources struct {
		Licenses  string `scfg:"licenses"`
		Static    string `scfg:"static"`
		Templates string `scfg:"templates"`
	} `scfg:"resources"`
	DB struct {
		Type string `scfg:"type"`
		Conn string `scfg:"conn"`
	} `scfg:"db"`
	Pprof struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
	} `scfg:"pprof"`
}

// Load reads the configuration file from the given path and unmarshals it into
// a Config value.
func Load(path string) (Config, error) {
	var cfg Config

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	decoder := scfg.NewDecoder(bufio.NewReader(f))
	if err = decoder.Decode(&cfg); err != nil {
		return cfg, err
	}
	for _, u := range decoder.UnknownDirectives() {
		slog.Warn("unknown configuration directive", "directive", u)
	}

	return cfg, nil
}
