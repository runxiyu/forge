// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

// global_data is passed as "global" when rendering HTML templates.
var global_data = map[string]any{
	"server_public_key_string":      &server_public_key_string,
	"server_public_key_fingerprint": &server_public_key_fingerprint,
	"forge_version":                 VERSION,
	// Some other ones are populated after config parsing
}
