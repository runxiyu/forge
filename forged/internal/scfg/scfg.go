// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

// Package scfg parses and formats configuration files.
// Note that this fork of scfg behaves differently from upstream scfg.
package scfg

import (
	"fmt"
)

// Block is a list of directives.
type Block []*Directive

// GetAll returns a list of directives with the provided name.
func (blk Block) GetAll(name string) []*Directive {
	l := make([]*Directive, 0, len(blk))
	for _, child := range blk {
		if child.Name == name {
			l = append(l, child)
		}
	}
	return l
}

// Get returns the first directive with the provided name.
func (blk Block) Get(name string) *Directive {
	for _, child := range blk {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// Directive is a configuration directive.
type Directive struct {
	Name   string
	Params []string

	Children Block

	lineno int
}

// ParseParams extracts parameters from the directive. It errors out if the
// user hasn't provided enough parameters.
func (d *Directive) ParseParams(params ...*string) error {
	if len(d.Params) < len(params) {
		return fmt.Errorf("directive %q: want %v params, got %v", d.Name, len(params), len(d.Params))
	}
	for i, ptr := range params {
		if ptr == nil {
			continue
		}
		*ptr = d.Params[i]
	}
	return nil
}
