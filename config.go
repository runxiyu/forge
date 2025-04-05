// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.lindenii.runxiyu.org/forge/internal/scfg"
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
	IRC struct {
		Net   string `scfg:"net"`
		Addr  string `scfg:"addr"`
		TLS   bool   `scfg:"tls"`
		SendQ uint   `scfg:"sendq"`
		Nick  string `scfg:"nick"`
		User  string `scfg:"user"`
		Gecos string `scfg:"gecos"`
	} `scfg:"irc"`
	General struct {
		Title string `scfg:"title"`
	} `scfg:"general"`
	DB struct {
		Type string `scfg:"type"`
		Conn string `scfg:"conn"`
	} `scfg:"db"`
}

// LoadConfig loads a configuration file from the specified path and unmarshals
// it to the global [config] struct. This may race with concurrent reads from
// [config]; additional synchronization is necessary if the configuration is to
// be made reloadable.
func (s *Server) LoadConfig(path string) (err error) {
	var configFile *os.File
	if configFile, err = os.Open(path); err != nil {
		return err
	}
	defer configFile.Close()

	decoder := scfg.NewDecoder(bufio.NewReader(configFile))
	if err = decoder.Decode(&s.config); err != nil {
		return err
	}
	for _, u := range decoder.UnknownDirectives() {
		slog.Warn("unknown configuration directive", "directive", u)
	}

	if s.config.DB.Type != "postgres" {
		return errors.New("unsupported database type")
	}

	if s.database, err = pgxpool.New(context.Background(), s.config.DB.Conn); err != nil {
		return err
	}

	s.globalData["forge_title"] = s.config.General.Title

	return nil
}
