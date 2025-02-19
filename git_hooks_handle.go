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
	"strings"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	err_get_fd    = errors.New("unable to get file descriptor")
	err_get_ucred = errors.New("failed getsockopt")
)

// hooks_handle_connection handles a connection from git_hooks_client via the
// unix socket.
func hooks_handle_connection(conn net.Conn) {
	defer conn.Close()

	// There aren't reasonable cases where someone would run this as
	// another user.
	ucred, err := get_ucred(conn)
	if err != nil {
		if _, err := conn.Write([]byte{1}); err != nil {
			return
		}
		fmt.Fprintln(conn, "Unable to get peer credentials:", err.Error())
		return
	}
	if ucred.Uid != uint32(os.Getuid()) {
		if _, err := conn.Write([]byte{1}); err != nil {
			return
		}
		fmt.Fprintln(conn, "UID mismatch")
		return
	}

	cookie := make([]byte, 64)
	_, err = conn.Read(cookie)
	if err != nil {
		if _, err := conn.Write([]byte{1}); err != nil {
			return
		}
		fmt.Fprintln(conn, "Failed to read cookie:", err.Error())
		return
	}

	pack_to_hook, ok := pack_to_hook_by_cookie.Load(string(cookie))
	if !ok {
		if _, err := conn.Write([]byte{1}); err != nil {
			return
		}
		fmt.Fprintln(conn, "Invalid handler cookie")
		return
	}

	ssh_stderr := pack_to_hook.session.Stderr()

	hook_return_value := func() byte {
		var argc64 uint64
		err = binary.Read(conn, binary.NativeEndian, &argc64)
		if err != nil {
			fmt.Fprintln(ssh_stderr, "Failed to read argc:", err.Error())
			return 1
		}
		var args []string
		for i := uint64(0); i < argc64; i++ {
			var arg bytes.Buffer
			for {
				b := make([]byte, 1)
				n, err := conn.Read(b)
				if err != nil || n != 1 {
					fmt.Fprintln(ssh_stderr, "Failed to read arg:", err.Error())
					return 1
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
				return 0
			} else {
				all_ok := true
				for {
					line, err := stdin.ReadString('\n')
					if errors.Is(err, io.EOF) {
						break
					}
					line = line[:len(line)-1]

					old_oid, rest, found := strings.Cut(line, " ")
					if !found {
						fmt.Fprintln(ssh_stderr, "Invalid pre-receive line:", line)
						return 1
					}

					new_oid, ref_name, found := strings.Cut(rest, " ")
					if !found {
						fmt.Fprintln(ssh_stderr, "Invalid pre-receive line:", line)
						return 1
					}

					if strings.HasPrefix(ref_name, "refs/heads/contrib/") {
						if all_zero_num_string(old_oid) { // New branch
							fmt.Fprintln(ssh_stderr, "Acceptable push to new contrib branch: "+ref_name)
							// TODO: Create a merge request. If that fails,
							// we should just reject this entire push
							// immediately.
						} else { // Existing contrib branch
							// TODO: Check if the current user is authorized
							// to push to this contrib branch.
							repo, err := git.PlainOpen(pack_to_hook.repo_path)
							if err != nil {
								fmt.Fprintln(ssh_stderr, "Daemon failed to open repo:", err.Error())
								return 1
							}

							old_hash := plumbing.NewHash(old_oid)

							old_commit, err := repo.CommitObject(old_hash)
							if err != nil {
								fmt.Fprintln(ssh_stderr, "Daemon failed to get old commit:", err.Error())
								return 1
							}

							// Potential BUG: I'm not sure if new_commit is guaranteed to be
							// detectable as they haven't been merged into the main repo's
							// objects yet. But it seems to work, and I don't think there's
							// any reason for this to only work intermitently.
							new_hash := plumbing.NewHash(new_oid)
							new_commit, err := repo.CommitObject(new_hash)
							if err != nil {
								fmt.Fprintln(ssh_stderr, "Daemon failed to get new commit:", err.Error())
								return 1
							}

							is_ancestor, err := old_commit.IsAncestor(new_commit)
							if err != nil {
								fmt.Fprintln(ssh_stderr, "Daemon failed to check if old commit is ancestor:", err.Error())
								return 1
							}

							if !is_ancestor {
								// TODO: Create MR snapshot ref instead
								all_ok = false
								fmt.Fprintln(ssh_stderr, "Rejecting force push to contrib branch: "+ref_name)
								continue
							}

							fmt.Fprintln(ssh_stderr, "Acceptable push to existing contrib branch: "+ref_name)
						}
					} else { // Non-contrib branch
						all_ok = false
						fmt.Fprintln(ssh_stderr, "Rejecting push to non-contrib branch: "+ref_name)
					}
				}

				if all_ok {
					return 0
				} else {
					return 1
				}
			}
		default:
			fmt.Fprintln(ssh_stderr, "Invalid hook:", args[0])
			return 1
		}
	}()

	_, _ = conn.Write([]byte{hook_return_value})
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

func all_zero_num_string(s string) bool {
	for _, r := range s {
		if r != '0' {
			return false
		}
	}
	return true
}
