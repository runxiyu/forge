package misc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
)

func ListenUnixSocket(ctx context.Context, path string) (listener net.Listener, replaced bool, err error) {
	listenConfig := net.ListenConfig{} //exhaustruct:ignore
	listener, err = listenConfig.Listen(ctx, "unix", path)
	if errors.Is(err, syscall.EADDRINUSE) {
		replaced = true
		unlinkErr := syscall.Unlink(path)
		if unlinkErr != nil {
			return listener, false, fmt.Errorf("remove existing socket %q: %w", path, unlinkErr)
		}
		listener, err = listenConfig.Listen(ctx, "unix", path)
	}
	if err != nil {
		return listener, replaced, fmt.Errorf("listen on unix socket %q: %w", path, err)
	}
	return listener, replaced, nil
}

func Listen(ctx context.Context, net_, addr string) (listener net.Listener, err error) {
	if net_ == "unix" {
		listener, _, err = ListenUnixSocket(ctx, addr)
		if err != nil {
			return listener, fmt.Errorf("listen unix socket for web: %w", err)
		}
	} else {
		listenConfig := net.ListenConfig{} //exhaustruct:ignore
		listener, err = listenConfig.Listen(ctx, net_, addr)
		if err != nil {
			return listener, fmt.Errorf("listen %s for web: %w", net_, err)
		}
	}
	return listener, nil
}
