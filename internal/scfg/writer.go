// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

package scfg

import (
	"errors"
	"io"
	"strings"
)

var errDirEmptyName = errors.New("scfg: directive with empty name")

// Write writes a parsed configuration to the provided io.Writer.
func Write(w io.Writer, blk Block) error {
	enc := newEncoder(w)
	err := enc.encodeBlock(blk)
	return err
}

// encoder write SCFG directives to an output stream.
type encoder struct {
	w   io.Writer
	lvl int
	err error
}

// newEncoder returns a new encoder that writes to w.
func newEncoder(w io.Writer) *encoder {
	return &encoder{w: w}
}

func (enc *encoder) push() {
	enc.lvl++
}

func (enc *encoder) pop() {
	enc.lvl--
}

func (enc *encoder) writeIndent() {
	for i := 0; i < enc.lvl; i++ {
		enc.write([]byte("\t"))
	}
}

func (enc *encoder) write(p []byte) {
	if enc.err != nil {
		return
	}
	_, enc.err = enc.w.Write(p)
}

func (enc *encoder) encodeBlock(blk Block) error {
	for _, dir := range blk {
		if err := enc.encodeDir(*dir); err != nil {
			return err
		}
	}
	return enc.err
}

func (enc *encoder) encodeDir(dir Directive) error {
	if enc.err != nil {
		return enc.err
	}

	if dir.Name == "" {
		enc.err = errDirEmptyName
		return enc.err
	}

	enc.writeIndent()
	enc.write([]byte(maybeQuote(dir.Name)))
	for _, p := range dir.Params {
		enc.write([]byte(" "))
		enc.write([]byte(maybeQuote(p)))
	}

	if len(dir.Children) > 0 {
		enc.write([]byte(" {\n"))
		enc.push()
		if err := enc.encodeBlock(dir.Children); err != nil {
			return err
		}
		enc.pop()

		enc.writeIndent()
		enc.write([]byte("}"))
	}
	enc.write([]byte("\n"))

	return enc.err
}

const specialChars = "\"\\\r\n'{} \t"

func maybeQuote(s string) string {
	if s == "" || strings.ContainsAny(s, specialChars) {
		var sb strings.Builder
		sb.WriteByte('"')
		for _, ch := range s {
			if strings.ContainsRune(`"\`, ch) {
				sb.WriteByte('\\')
			}
			sb.WriteRune(ch)
		}
		sb.WriteByte('"')
		return sb.String()
	}
	return s
}
