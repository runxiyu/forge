// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

// Package embed provides embedded filesystems created in build-time.
package embed

import "embed"

// Source contains the licenses collected at build time.
// It is intended to be served to the user.
//
//go:embed LICENSE*
var Source embed.FS

// Resources contains the templates and static files used by the web interface,
// as well as the git backend daemon and the hookc helper.
//
//go:embed forged/templates/* forged/static/*
//go:embed hookc/hookc git2d/git2d
var Resources embed.FS
