package embed

import "embed"

//go:embed LICENSE* source.tar.gz
var Source embed.FS

//go:embed templates/* static/*
//go:embed hookc/hookc git2d/git2d
var Resources embed.FS
