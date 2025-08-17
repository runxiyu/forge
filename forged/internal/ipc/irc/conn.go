package irc

import (
	"bufio"
	"fmt"
	"net"
	"slices"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
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
	n, err = c.netConn.Write(p)
	if err != nil {
		err = fmt.Errorf("write to connection: %w", err)
	}
	return n, err
}

func (c *Conn) WriteString(s string) (n int, err error) {
	n, err = c.netConn.Write(misc.StringToBytes(s))
	if err != nil {
		err = fmt.Errorf("write to connection: %w", err)
	}
	return n, err
}
