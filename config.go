// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bufio"
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.lindenii.runxiyu.org/lindenii-common/scfg"
)

var database *pgxpool.Pool

var err_unsupported_database_type = errors.New("unsupported database type")

var config struct {
	HTTP struct {
		Net          string `scfg:"net"`
		Addr         string `scfg:"addr"`
		CookieExpiry int    `scfg:"cookie_expiry"`
		Root         string `scfg:"root"`
	} `scfg:"http"`
	Hooks struct {
		Socket string `scfg:"socket"`
		Execs  string `scfg:"execs"`
	} `scfg:"hooks"`
	SSH struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
		Key  string `scfg:"key"`
		Root string `scfg:"root"`
	} `scfg:"ssh"`
	General struct {
		Title string `scfg:"title"`
	} `scfg:"general"`
	DB struct {
		Type string `scfg:"type"`
		Conn string `scfg:"conn"`
	} `scfg:"db"`
}

func load_config(path string) (err error) {
	var config_file *os.File
	var decoder *scfg.Decoder

	if config_file, err = os.Open(path); err != nil {
		return err
	}
	defer config_file.Close()

	decoder = scfg.NewDecoder(bufio.NewReader(config_file))
	if err = decoder.Decode(&config); err != nil {
		return err
	}

	if config.DB.Type != "postgres" {
		return err_unsupported_database_type
	}

	if database, err = pgxpool.New(context.Background(), config.DB.Conn); err != nil {
		return err
	}

	global_data["forge_title"] = config.General.Title

	return nil
}
