package main

import (
	"crypto/tls"
	"net"

	"go.lindenii.runxiyu.org/lindenii-irc"
)

func ircBotSession() error {
	var err error
	var underlyingConn net.Conn
	if config.IRC.TLS {
		underlyingConn, err = tls.Dial(config.IRC.Net, config.IRC.Addr, nil)
	} else {
		underlyingConn, err = net.Dial(config.IRC.Net, config.IRC.Addr)
	}
	if err != nil {
		return (err)
	}
	conn := irc.NewConn(underlyingConn)
	conn.WriteString("NICK forge\r\nUSER forge 0 * :Forge\r\n")
	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			return (err)
		}
		switch msg.Command {
		case "001":
			conn.WriteString("JOIN #chat\r\n")
		case "PING":
			conn.WriteString("PONG :")
			conn.WriteString(msg.Args[0])
			conn.WriteString("\r\n")
		case "JOIN":
			conn.WriteString("PRIVMSG #chat :test\r\n")
		default:
		}
	}
}

func ircBotLoop() {
	for {
		_ = ircBotSession()
	}
}
