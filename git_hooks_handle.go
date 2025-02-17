package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

var err_not_unixconn = errors.New("Not a unix connection")
var err_get_fd = errors.New("Unable to get file descriptor")
var err_get_ucred = errors.New("Failed getsockopt")

func hooks_handle_connection(conn net.Conn) {
	defer conn.Close()

	ucred, err := get_ucred(conn)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Unable to get peer credentials:", err.Error())
		return
	}

	if ucred.Uid != uint32(os.Getuid()) {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "UID mismatch")
		return
	}

	conn.Write([]byte{0})
	fmt.Fprintf(conn, "Your PID is %d\n", ucred.Pid)

	return
}

func serve_git_hooks(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go hooks_handle_connection(conn)
	}
}

func get_ucred(conn net.Conn) (*syscall.Ucred, error) {
	unix_conn := conn.(*net.UnixConn)
	fd, err := unix_conn.File()
	if err != nil {
		return nil, err_get_fd
	}
	defer fd.Close()

	ucred, err := syscall.GetsockoptUcred(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return nil, err_get_ucred
	}
	return ucred, nil
}
