package irc

import (
	"bufio"
	"net"
	"slices"

	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

type Conn struct {
	netConn   net.Conn
	bufReader *bufio.Reader
}

func NewConn(netConn net.Conn) Conn {
	return Conn{
		netConn:   netConn,
		bufReader: bufio.NewReader(netConn),
	}
}

func (c *Conn) ReadMessage() (msg Message, line string, err error) {
	raw, err := c.bufReader.ReadSlice('\n')
	if err != nil {
		return
	}

	if raw[len(raw)-1] == '\n' {
		raw = raw[:len(raw)-1]
	}
	if raw[len(raw)-1] == '\r' {
		raw = raw[:len(raw)-1]
	}

	lineBytes := slices.Clone(raw)
	line = misc.BytesToString(lineBytes)
	msg, err = Parse(lineBytes)

	return
}

func (c *Conn) Write(p []byte) (n int, err error) {
	return c.netConn.Write(p)
}

func (c *Conn) WriteString(s string) (n int, err error) {
	return c.netConn.Write(misc.StringToBytes(s))
}
