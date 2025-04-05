// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"flag"
	"log/slog"
	"os"

	"go.lindenii.runxiyu.org/forge"
)

func main() {
	configPath := flag.String(
		"config",
		"/etc/lindenii/forge.scfg",
		"path to configuration file",
	)
	flag.Parse()

	s := forge.Server{}

	s.Setup()

	if err := s.LoadConfig(*configPath); err != nil {
		slog.Error("loading configuration", "error", err)
		os.Exit(1)
	}

	s.Run()
}
