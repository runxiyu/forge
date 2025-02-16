package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

var err_not_unixconn = errors.New("Not a unix connection")

func hooks_handle_connection(conn net.Conn) (err error) {
	defer conn.Close()

	unix_conn, ok := conn.(*net.UnixConn)
	if !ok {
		return err_not_unixconn
	}

	fd, err := unix_conn.File()
	if err != nil {
		return err
	}
	defer fd.Close()

	ucred, err := get_ucred(fd)
	if err != nil {
		return err
	}

	pid := ucred.Pid

	_ = pid

	return nil
}

func get_ucred(fd *os.File) (*syscall.Ucred, error) {
	ucred, err := syscall.GetsockoptUcred(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %v", err)
	}
	return ucred, nil
}
