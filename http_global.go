// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

// globalData is passed as "global" when rendering HTML templates.
var globalData = map[string]any{
	"server_public_key_string":      &serverPubkeyString,
	"server_public_key_fingerprint": &serverPubkeyFP,
	"forge_version":                 VERSION,
	// Some other ones are populated after config parsing
}
