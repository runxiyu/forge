package misc

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

func ListenUnixSocket(path string) (listener net.Listener, replaced bool, err error) {
	listener, err = net.Listen("unix", path)
	if errors.Is(err, syscall.EADDRINUSE) {
		replaced = true
		if unlinkErr := syscall.Unlink(path); unlinkErr != nil {
			return listener, false, fmt.Errorf("remove existing socket %q: %w", path, unlinkErr)
		}
		listener, err = net.Listen("unix", path)
	}
	if err != nil {
		return listener, replaced, fmt.Errorf("listen on unix socket %q: %w", path, err)
	}
	return listener, replaced, nil
}

func Listen(net_, addr string) (listener net.Listener, err error) {
	if net_ == "unix" {
		listener, _, err = ListenUnixSocket(addr)
		if err != nil {
			return listener, fmt.Errorf("listen unix socket for web: %w", err)
		}
	} else {
		listener, err = net.Listen(net_, addr)
		if err != nil {
			return listener, fmt.Errorf("listen %s for web: %w", net_, err)
		}
	}
	return listener, nil
}
