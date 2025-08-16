// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package irc

import (
	"bytes"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
)

type Source interface {
	AsSourceString() string
}

func parseSource(s []byte) Source {
	nick, userhost, found := bytes.Cut(s, []byte{'!'})
	if !found {
		return Server{name: misc.BytesToString(s)}
	}

	user, host, found := bytes.Cut(userhost, []byte{'@'})
	if !found {
		return Server{name: misc.BytesToString(s)}
	}

	return Client{
		Nick: misc.BytesToString(nick),
		User: misc.BytesToString(user),
		Host: misc.BytesToString(host),
	}
}

type Server struct {
	name string
}

func (s Server) AsSourceString() string {
	return s.name
}

type Client struct {
	Nick string
	User string
	Host string
}

func (c Client) AsSourceString() string {
	return c.Nick + "!" + c.User + "@" + c.Host
}
