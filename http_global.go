// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

// globalData is passed as "global" when rendering HTML templates and contains
// global data that should stay constant throughout an execution of Lindenii
// Forge as no synchronization mechanism is provided for updating it.
var globalData = map[string]any{
	"server_public_key_string":      &serverPubkeyString,
	"server_public_key_fingerprint": &serverPubkeyFP,
	"forge_version":                 VERSION,
	// Some other ones are populated after config parsing
}
