// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2020 Simon Ser <https://emersion.fr>

package scfg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// This limits the max block nesting depth to prevent stack overflows.
const maxNestingDepth = 1000

// Load loads a configuration file.
func Load(path string) (block Block, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	return Read(f)
}

// Read parses a configuration file from an io.Reader.
func Read(r io.Reader) (Block, error) {
	scanner := bufio.NewScanner(r)

	dec := decoder{scanner: scanner}
	block, closingBrace, err := dec.readBlock()
	if err != nil {
		return nil, err
	} else if closingBrace {
		return nil, fmt.Errorf("line %v: unexpected '}'", dec.lineno)
	}

	return block, scanner.Err()
}

type decoder struct {
	scanner    *bufio.Scanner
	lineno     int
	blockDepth int
}

// readBlock reads a block. closingBrace is true if parsing stopped on '}'
// (otherwise, it stopped on Scanner.Scan).
func (dec *decoder) readBlock() (block Block, closingBrace bool, err error) {
	dec.blockDepth++
	defer func() {
		dec.blockDepth--
	}()

	if dec.blockDepth >= maxNestingDepth {
		return nil, false, fmt.Errorf("exceeded max block depth")
	}

	for dec.scanner.Scan() {
		dec.lineno++

		l := dec.scanner.Text()
		words, err := splitWords(l)
		if err != nil {
			return nil, false, fmt.Errorf("line %v: %v", dec.lineno, err)
		} else if len(words) == 0 {
			continue
		}

		if len(words) == 1 && l[len(l)-1] == '}' {
			closingBrace = true
			break
		}

		var d *Directive
		if words[len(words)-1] == "{" && l[len(l)-1] == '{' {
			words = words[:len(words)-1]

			var name string
			params := words
			if len(words) > 0 {
				name, params = words[0], words[1:]
			}

			startLineno := dec.lineno
			childBlock, childClosingBrace, err := dec.readBlock()
			if err != nil {
				return nil, false, err
			} else if !childClosingBrace {
				return nil, false, fmt.Errorf("line %v: unterminated block", startLineno)
			}

			// Allows callers to tell apart "no block" and "empty block"
			if childBlock == nil {
				childBlock = Block{}
			}

			d = &Directive{Name: name, Params: params, Children: childBlock, lineno: dec.lineno}
		} else {
			d = &Directive{Name: words[0], Params: words[1:], lineno: dec.lineno}
		}
		block = append(block, d)
	}

	return block, closingBrace, nil
}

func splitWords(l string) ([]string, error) {
	var (
		words   []string
		sb      strings.Builder
		escape  bool
		quote   rune
		wantWSP bool
	)
	for _, ch := range l {
		switch {
		case escape:
			sb.WriteRune(ch)
			escape = false
		case wantWSP && (ch != ' ' && ch != '\t'):
			return words, fmt.Errorf("atom not allowed after quoted string")
		case ch == '\\':
			escape = true
		case quote != 0 && ch == quote:
			quote = 0
			wantWSP = true
			if sb.Len() == 0 {
				words = append(words, "")
			}
		case quote == 0 && len(words) == 0 && sb.Len() == 0 && ch == '#':
			return nil, nil
		case quote == 0 && (ch == '\'' || ch == '"'):
			if sb.Len() > 0 {
				return words, fmt.Errorf("quoted string not allowed after atom")
			}
			quote = ch
		case quote == 0 && (ch == ' ' || ch == '\t'):
			if sb.Len() > 0 {
				words = append(words, sb.String())
			}
			sb.Reset()
			wantWSP = false
		default:
			sb.WriteRune(ch)
		}
	}
	if quote != 0 {
		return words, fmt.Errorf("unterminated quoted string")
	}
	if sb.Len() > 0 {
		words = append(words, sb.String())
	}
	return words, nil
}
