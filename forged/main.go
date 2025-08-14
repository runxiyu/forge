// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// The main entry point to the Lindenii Forge daemon.
package main

import (
	"context"
	"flag"

	"go.lindenii.runxiyu.org/forge/forged/internal/server"
)

func main() {
	configPath := flag.String(
		"config",
		"/etc/lindenii/forge.scfg",
		"path to configuration file",
	)
	flag.Parse()

	s, err := server.New(context.Background(), *configPath)
	if err != nil {
		panic(err)
	}

	panic(s.Run())
}
