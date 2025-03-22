package main

import (
	"crypto/tls"
	"net"

	"go.lindenii.runxiyu.org/lindenii-common/clog"
	irc "go.lindenii.runxiyu.org/lindenii-irc"
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

	logAndWriteLn := func(s string) (n int, err error) {
		clog.Debug("IRC tx: " + s)
		return conn.WriteString(s + "\r\n")
	}

	_, err = logAndWriteLn("NICK " + config.IRC.Nick)
	if err != nil {
		return err
	}
	_, err = logAndWriteLn("USER " + config.IRC.User + " 0 * :" + config.IRC.Gecos)
	if err != nil {
		return err
	}

	readLoopError := make(chan error)
	writeLoopAbort := make(chan struct{})
	go func() {
		for {
			select {
			case <-writeLoopAbort:
				return
			default:
			}

			msg, line, err := conn.ReadMessage()
			if err != nil {
				readLoopError <- err
				return
			}

			clog.Debug("IRC rx: " + line)

			switch msg.Command {
			case "001":
				_, err = logAndWriteLn("JOIN #chat")
				if err != nil {
					readLoopError <- err
					return
				}
			case "PING":
				_, err = logAndWriteLn("PONG :" + msg.Args[0])
				if err != nil {
					readLoopError <- err
					return
				}
			case "JOIN":
				c, ok := msg.Source.(irc.Client)
				if !ok {
					clog.Error("IRC server told us a non-client is joining a channel...")
				}
				if c.Nick != config.IRC.Nick {
					continue
				}
				_, err = logAndWriteLn("PRIVMSG #chat :test")
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
		case line := <-ircSendBuffered:
			_, err = logAndWriteLn(line)
			if err != nil {
				select {
				case ircSendBuffered <- line:
				default:
					clog.Error("unable to requeue IRC message: " + line)
				}
				writeLoopAbort <- struct{}{}
				return err
			}
		case lineErrorBack := <-ircSendDirectChan:
			_, err = logAndWriteLn(lineErrorBack.content)
			lineErrorBack.errorBack <- err
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
