// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/lindenii-common/ansiec"
)

var (
	err_get_fd    = errors.New("unable to get file descriptor")
	err_get_ucred = errors.New("failed getsockopt")
)

// hooks_handle_connection handles a connection from git_hooks_client via the
// unix socket.
func hooks_handle_connection(conn net.Conn) {
	var ctx context.Context
	var cancel context.CancelFunc
	var ucred *syscall.Ucred
	var err error
	var cookie []byte
	var pack_to_hook pack_to_hook_t
	var ssh_stderr io.Writer
	var ok bool
	var hook_return_value byte

	defer conn.Close()
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// There aren't reasonable cases where someone would run this as
	// another user.
	if ucred, err = get_ucred(conn); err != nil {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		wf_error(conn, "\nUnable to get peer credentials: %v", err)
		return
	}
	if ucred.Uid != uint32(os.Getuid()) {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		wf_error(conn, "\nUID mismatch")
		return
	}

	cookie = make([]byte, 64)
	if _, err = conn.Read(cookie); err != nil {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		wf_error(conn, "\nFailed to read cookie: %v", err)
		return
	}

	pack_to_hook, ok = pack_to_hook_by_cookie.Load(string(cookie))
	if !ok {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		wf_error(conn, "\nInvalid handler cookie")
		return
	}

	ssh_stderr = pack_to_hook.session.Stderr()

	_, _ = ssh_stderr.Write([]byte{'\n'})

	hook_return_value = func() byte {
		var argc64 uint64
		if err = binary.Read(conn, binary.NativeEndian, &argc64); err != nil {
			wf_error(ssh_stderr, "Failed to read argc: %v", err)
			return 1
		}
		var args []string
		for i := uint64(0); i < argc64; i++ {
			var arg bytes.Buffer
			for {
				b := make([]byte, 1)
				n, err := conn.Read(b)
				if err != nil || n != 1 {
					wf_error(ssh_stderr, "Failed to read arg: %v", err)
					return 1
				}
				if b[0] == 0 {
					break
				}
				arg.WriteByte(b[0])
			}
			args = append(args, arg.String())
		}

		git_env := make(map[string]string)
		for {
			var env_line bytes.Buffer
			for {
				b := make([]byte, 1)
				n, err := conn.Read(b)
				if err != nil || n != 1 {
					wf_error(ssh_stderr, "Failed to read environment variable: %v", err)
					return 1
				}
				if b[0] == 0 {
					break
				}
				env_line.WriteByte(b[0])
			}
			if env_line.Len() == 0 {
				break
			}
			kv := env_line.String()
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) < 2 {
				wf_error(ssh_stderr, "Invalid environment variable line: %v", kv)
				return 1
			}
			git_env[parts[0]] = parts[1]
		}

		fmt.Printf("%#v\n", git_env)

		var stdin bytes.Buffer
		if _, err = io.Copy(&stdin, conn); err != nil {
			wf_error(conn, "Failed to read to the stdin buffer: %v", err)
		}

		switch filepath.Base(args[0]) {
		case "pre-receive":
			if pack_to_hook.direct_access {
				return 0
			} else {
				all_ok := true
				for {
					var line, old_oid, rest, new_oid, ref_name string
					var found bool
					var old_hash, new_hash plumbing.Hash
					var old_commit, new_commit *object.Commit

					line, err = stdin.ReadString('\n')
					if errors.Is(err, io.EOF) {
						break
					} else if err != nil {
						wf_error(ssh_stderr, "Failed to read pre-receive line: %v", err)
						return 1
					}
					line = line[:len(line)-1]

					old_oid, rest, found = strings.Cut(line, " ")
					if !found {
						wf_error(ssh_stderr, "Invalid pre-receive line: %v", line)
						return 1
					}

					new_oid, ref_name, found = strings.Cut(rest, " ")
					if !found {
						wf_error(ssh_stderr, "Invalid pre-receive line: %v", line)
						return 1
					}

					if strings.HasPrefix(ref_name, "refs/heads/contrib/") {
						if all_zero_num_string(old_oid) { // New branch
							fmt.Fprintln(ssh_stderr, ansiec.Blue+"POK"+ansiec.Reset, ref_name)
							var new_mr_id int

							err = database.QueryRow(ctx,
								"INSERT INTO merge_requests (repo_id, creator, source_ref, status) VALUES ($1, $2, $3, 'open') RETURNING id",
								pack_to_hook.repo_id, pack_to_hook.user_id, strings.TrimPrefix(ref_name, "refs/heads/"),
							).Scan(&new_mr_id)
							if err != nil {
								wf_error(ssh_stderr, "Error creating merge request: %v", err)
								return 1
							}
							fmt.Fprintln(ssh_stderr, ansiec.Blue+"Created merge request at", generate_http_remote_url(pack_to_hook.group_path, pack_to_hook.repo_name)+"/contrib/"+strconv.FormatUint(uint64(new_mr_id), 10)+"/"+ansiec.Reset)
						} else { // Existing contrib branch
							var existing_merge_request_user_id int
							var is_ancestor bool

							err = database.QueryRow(ctx,
								"SELECT COALESCE(creator, 0) FROM merge_requests WHERE source_ref = $1 AND repo_id = $2",
								strings.TrimPrefix(ref_name, "refs/heads/"), pack_to_hook.repo_id,
							).Scan(&existing_merge_request_user_id)
							if err != nil {
								if errors.Is(err, pgx.ErrNoRows) {
									wf_error(ssh_stderr, "No existing merge request for existing contrib branch: %v", err)
								} else {
									wf_error(ssh_stderr, "Error querying for existing merge request: %v", err)
								}
								return 1
							}
							if existing_merge_request_user_id == 0 {
								all_ok = false
								fmt.Fprintln(ssh_stderr, ansiec.Red+"NAK"+ansiec.Reset, ref_name, "(branch belongs to unowned MR)")
								continue
							}

							if existing_merge_request_user_id != pack_to_hook.user_id {
								all_ok = false
								fmt.Fprintln(ssh_stderr, ansiec.Red+"NAK"+ansiec.Reset, ref_name, "(branch belongs another user's MR)")
								continue
							}

							old_hash = plumbing.NewHash(old_oid)

							if old_commit, err = pack_to_hook.repo.CommitObject(old_hash); err != nil {
								wf_error(ssh_stderr, "Daemon failed to get old commit: %v", err)
								return 1
							}

							// Potential BUG: I'm not sure if new_commit is guaranteed to be
							// detectable as they haven't been merged into the main repo's
							// objects yet. But it seems to work, and I don't think there's
							// any reason for this to only work intermitently.
							new_hash = plumbing.NewHash(new_oid)
							if new_commit, err = pack_to_hook.repo.CommitObject(new_hash); err != nil {
								wf_error(ssh_stderr, "Daemon failed to get new commit: %v", err)
								return 1
							}

							if is_ancestor, err = old_commit.IsAncestor(new_commit); err != nil {
								wf_error(ssh_stderr, "Daemon failed to check if old commit is ancestor: %v", err)
								return 1
							}

							if !is_ancestor {
								// TODO: Create MR snapshot ref instead
								all_ok = false
								fmt.Fprintln(ssh_stderr, ansiec.Red+"NAK"+ansiec.Reset, ref_name, "(force pushes are not supported yet)")
								continue
							}

							fmt.Fprintln(ssh_stderr, ansiec.Blue+"POK"+ansiec.Reset, ref_name)
						}
					} else { // Non-contrib branch
						all_ok = false
						fmt.Fprintln(ssh_stderr, ansiec.Red+"NAK"+ansiec.Reset, ref_name, "(you cannot push to branches outside of contrib/*)")
					}
				}

				fmt.Fprintln(ssh_stderr)
				if all_ok {
					fmt.Fprintln(ssh_stderr, "Overall "+ansiec.Green+"ACK"+ansiec.Reset+" (all checks passed)")
					return 0
				} else {
					fmt.Fprintln(ssh_stderr, "Overall "+ansiec.Red+"NAK"+ansiec.Reset+" (one or more branches failed checks)")
					return 1
				}
			}
		default:
			fmt.Fprintln(ssh_stderr, ansiec.Red+"Invalid hook:", args[0]+ansiec.Reset)
			return 1
		}
	}()

	fmt.Fprintln(ssh_stderr)

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

func get_ucred(conn net.Conn) (ucred *syscall.Ucred, err error) {
	var unix_conn *net.UnixConn = conn.(*net.UnixConn)
	var fd *os.File

	if fd, err = unix_conn.File(); err != nil {
		return nil, err_get_fd
	}
	defer fd.Close()

	if ucred, err = syscall.GetsockoptUcred(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED); err != nil {
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
