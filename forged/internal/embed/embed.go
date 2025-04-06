// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package embed provides embedded filesystems created in build-time.
package embed

import "embed"

//go:embed LICENSE* source.tar.gz
var Source embed.FS

//go:embed forged/templates/* forged/static/*
//go:embed hookc/hookc git2d/git2d
var Resources embed.FS
