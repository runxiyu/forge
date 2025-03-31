// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"bufio"
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.lindenii.runxiyu.org/lindenii-common/scfg"
)

// config holds the global configuration used by this instance. There is
// currently no synchronization mechanism, so it must not be modified after
// request handlers are spawned.
var config struct {
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
		Socket string `scfg:"socket"`
	} `scfg:"lmtp"`
	Git struct {
		RepoDir string `scfg:"repo_dir"`
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

// loadConfig loads a configuration file from the specified path and unmarshals
// it to the global [config] struct. This may race with concurrent reads from
// [config]; additional synchronization is necessary if the configuration is to
// be made reloadable.
//
// TODO: Currently, it returns an error when the user specifies any unknown
// configuration patterns, but silently ignores fields in the [config] struct
// that is not present in the user's configuration file. We would prefer the
// exact opposite behavior.
func loadConfig(path string) (err error) {
	var configFile *os.File
	if configFile, err = os.Open(path); err != nil {
		return err
	}
	defer configFile.Close()

	decoder := scfg.NewDecoder(bufio.NewReader(configFile))
	if err = decoder.Decode(&config); err != nil {
		return err
	}

	if config.DB.Type != "postgres" {
		return errors.New("unsupported database type")
	}

	if database, err = pgxpool.New(context.Background(), config.DB.Conn); err != nil {
		return err
	}

	globalData["forge_title"] = config.General.Title

	return nil
}
