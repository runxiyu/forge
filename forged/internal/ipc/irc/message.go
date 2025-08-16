// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2018-2024 luk3yx <https://luk3yx.github.io>
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package irc

import (
	"bytes"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Message struct {
	Command string
	Source  Source
	Tags    map[string]string
	Args    []string
}

// All strings returned are borrowed from the input byte slice.
func Parse(raw []byte) (msg Message, err error) {
	sp := bytes.Split(raw, []byte{' '}) // TODO: Use bytes.Cut instead here

	if bytes.HasPrefix(sp[0], []byte{'@'}) { // TODO: Check size manually
		if len(sp[0]) < 2 {
			err = ErrMalformedMsg
			return
		}
		sp[0] = sp[0][1:]

		msg.Tags, err = tagsToMap(sp[0])
		if err != nil {
			return
		}

		if len(sp) < 2 {
			err = ErrMalformedMsg
			return
		}
		sp = sp[1:]
	} else {
		msg.Tags = nil // TODO: Is a nil map the correct thing to use here?
	}

	if bytes.HasPrefix(sp[0], []byte{':'}) { // TODO: Check size manually
		if len(sp[0]) < 2 {
			err = ErrMalformedMsg
			return
		}
		sp[0] = sp[0][1:]

		msg.Source = parseSource(sp[0])

		if len(sp) < 2 {
			err = ErrMalformedMsg
			return
		}
		sp = sp[1:]
	}

	msg.Command = misc.BytesToString(sp[0])
	if len(sp) < 2 {
		return
	}
	sp = sp[1:]

	for i := 0; i < len(sp); i++ {
		if len(sp[i]) == 0 {
			continue
		}
		if sp[i][0] == ':' {
			if len(sp[i]) < 2 {
				sp[i] = []byte{}
			} else {
				sp[i] = sp[i][1:]
			}
			msg.Args = append(msg.Args, misc.BytesToString(bytes.Join(sp[i:], []byte{' '})))
			// TODO: Avoid Join by not using sp in the first place
			break
		}
		msg.Args = append(msg.Args, misc.BytesToString(sp[i]))
	}

	return
}

var ircv3TagEscapes = map[byte]byte{ //nolint:gochecknoglobals
	':': ';',
	's': ' ',
	'r': '\r',
	'n': '\n',
}

func tagsToMap(raw []byte) (tags map[string]string, err error) {
	tags = make(map[string]string)
	for rawTag := range bytes.SplitSeq(raw, []byte{';'}) {
		key, value, found := bytes.Cut(rawTag, []byte{'='})
		if !found {
			err = ErrInvalidIRCv3Tag
			return
		}
		if len(value) == 0 {
			tags[misc.BytesToString(key)] = ""
		} else {
			if !bytes.Contains(value, []byte{'\\'}) {
				tags[misc.BytesToString(key)] = misc.BytesToString(value)
			} else {
				valueUnescaped := bytes.NewBuffer(make([]byte, 0, len(value)))
				for i := 0; i < len(value); i++ {
					if value[i] == '\\' {
						i++
						byteUnescaped, ok := ircv3TagEscapes[value[i]]
						if !ok {
							byteUnescaped = value[i]
						}
						valueUnescaped.WriteByte(byteUnescaped)
					} else {
						valueUnescaped.WriteByte(value[i])
					}
				}
				tags[misc.BytesToString(key)] = misc.BytesToString(valueUnescaped.Bytes())
			}
		}
	}
	return
}
