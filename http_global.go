package main

var global_data = map[string]any{
	"server_public_key_string":      &server_public_key_string,
	"server_public_key_fingerprint": &server_public_key_fingerprint,
	// Some other ones are populated after config parsing
}
