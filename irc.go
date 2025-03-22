package main

import (
	"crypto/tls"
	"net"

	"go.lindenii.runxiyu.org/lindenii-common/clog"
	"go.lindenii.runxiyu.org/lindenii-irc"
)

var (
	ircSendBuffered   chan string
	ircSendDirectChan chan errorBack[string]
)

type errorBack[T any] struct {
	content   T
	errorBack chan error
}

func ircBotSession() error {
	var err error
	var underlyingConn net.Conn
	if config.IRC.TLS {
		underlyingConn, err = tls.Dial(config.IRC.Net, config.IRC.Addr, nil)
	} else {
		underlyingConn, err = net.Dial(config.IRC.Net, config.IRC.Addr)
	}
	if err != nil {
		return err
	}
	defer underlyingConn.Close()

	conn := irc.NewConn(underlyingConn)
	conn.WriteString(
		"NICK " + config.IRC.Nick + "\r\n" +
			"USER " + config.IRC.User + " 0 * :" + config.IRC.Gecos + "\r\n",
	)

	readLoopError := make(chan error)
	writeLoopAbort := make(chan struct{})
	go func() {
		for {
			select {
			case <-writeLoopAbort:
				return
			default:
			}
			msg, err := conn.ReadMessage()
			if err != nil {
				readLoopError <- err
				return
			}
			switch msg.Command {
			case "001":
				_, err = conn.WriteString("JOIN #chat\r\n")
				if err != nil {
					readLoopError <- err
					return
				}
			case "PING":
				_, err = conn.WriteString("PONG :" + msg.Args[0] + "\r\n")
				if err != nil {
					readLoopError <- err
					return
				}
			case "JOIN":
				_, err = conn.WriteString("PRIVMSG #chat :test\r\n")
				if err != nil {
					readLoopError <- err
					return
				}
			default:
			}
		}
	}()

	for {
		select {
		case err = <-readLoopError:
			return err
		case s := <-ircSendBuffered:
			_, err = conn.WriteString(s)
			if err != nil {
				select {
				case ircSendBuffered <- s:
				default:
					clog.Error("unable to requeue IRC message: " + s)
				}
				writeLoopAbort <- struct{}{}
				return err
			}
		case se := <-ircSendDirectChan:
			_, err = conn.WriteString(se.content)
			se.errorBack <- err
			if err != nil {
				writeLoopAbort <- struct{}{}
				return err
			}
		}
	}
}

func ircSendDirect(s string) error {
	ech := make(chan error, 1)

	ircSendDirectChan <- errorBack[string]{
		content:   s,
		errorBack: ech,
	}

	return <-ech
}

func ircBotLoop() {
	ircSendBuffered = make(chan string, config.IRC.SendQ)
	ircSendDirectChan = make(chan errorBack[string])

	for {
		err := ircBotSession()
		clog.Error("IRC error: " + err.Error())
	}
}
