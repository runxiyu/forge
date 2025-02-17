package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

var err_not_unixconn = errors.New("Not a unix connection")

func hooks_handle_connection(conn net.Conn) {
	defer conn.Close()

	unix_conn := conn.(*net.UnixConn)

	fd, err := unix_conn.File()
	if err != nil {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Unable to get file descriptor")
		return
	}
	defer fd.Close()

	ucred, err := get_ucred(fd)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Unable to get peer credentials")
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

func get_ucred(fd *os.File) (*syscall.Ucred, error) {
	ucred, err := syscall.GetsockoptUcred(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %v", err)
	}
	return ucred, nil
}
