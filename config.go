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

var err_unsupported_database_type = errors.New("Unsupported database type")

var config struct {
	HTTP struct {
		Net          string `scfg:"net"`
		Addr         string `scfg:"addr"`
		CookieExpiry int    `scfg:"cookie_expiry"`
	} `scfg:"http"`
	SSH struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
		Key  string `scfg:"key"`
	} `scfg:"ssh"`
	Git struct {
		Root string `scfg:"root"`
	} `scfg:"git"`
	DB struct {
		Type string `scfg:"type"`
		Conn string `scfg:"conn"`
	} `scfg:"db"`
}

func load_config(path string) (err error) {
	config_file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer config_file.Close()

	decoder := scfg.NewDecoder(bufio.NewReader(config_file))
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	if config.DB.Type != "postgres" {
		return err_unsupported_database_type
	}
	database, err = pgxpool.New(context.Background(), config.DB.Conn)
	if err != nil {
		return err
	}

	return nil
}
