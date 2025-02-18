package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
)

var (
	err_get_fd    = errors.New("Unable to get file descriptor")
	err_get_ucred = errors.New("Failed getsockopt")
)

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

	cookie := make([]byte, 64)
	_, err = conn.Read(cookie)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Failed to read cookie:", err.Error())
		return
	}

	pack_to_hook, ok := pack_to_hook_by_cookie.Load(string(cookie))
	if !ok {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Invalid handler cookie")
		return
	}

	var argc64 uint64
	err = binary.Read(conn, binary.NativeEndian, &argc64)
	if err != nil {
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Failed to read argc:", err.Error())
		return
	}
	var args []string
	for i := uint64(0); i < argc64; i++ {
		var arg bytes.Buffer
		for {
			b := make([]byte, 1)
			n, err := conn.Read(b)
			if err != nil || n != 1 {
				conn.Write([]byte{1})
				fmt.Fprintln(conn, "Failed to read arg:", err.Error())
				return
			}
			if b[0] == 0 {
				break
			}
			arg.WriteByte(b[0])
		}
		args = append(args, arg.String())
	}

	var stdin bytes.Buffer
	_, err = io.Copy(&stdin, conn)
	if err != nil {
		fmt.Fprintln(conn, "Failed to read to the stdin buffer:", err.Error())
	}

	switch filepath.Base(args[0]) {
	case "pre-receive":
		if pack_to_hook.direct_access {
			conn.Write([]byte{0})
		} else {
			conn.Write([]byte{1})
			fmt.Fprintln(conn, "Non-maintainer push access not implemented yet")
		}
	default:
		conn.Write([]byte{1})
		fmt.Fprintln(conn, "Invalid hook:", args[0])
	}
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
