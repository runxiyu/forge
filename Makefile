# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

.RECIPEPREFIX := >
.PHONY: forge clean

VERSION = $(shell git describe --tags --always --dirty)

forge:
>CGO_ENABLED=0 go build -o forge -ldflags '-X "go.lindenii.runxiyu.org/forge/forged/internal/server.version=$(VERSION)"' ./forged

clean:
>rm -rf forge
