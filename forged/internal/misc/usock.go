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
