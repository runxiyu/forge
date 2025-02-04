package main

import (
	"bufio"
	"os"

	"go.lindenii.runxiyu.org/lindenii-common/scfg"
)

var config struct {
	HTTP struct {
		Net  string `scfg:"net"`
		Addr string `scfg:"addr"`
	} `scfg:"http"`
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

	decoder := scfg.NewDecoder(bufio.NewReader(config_file))
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	return nil
}
