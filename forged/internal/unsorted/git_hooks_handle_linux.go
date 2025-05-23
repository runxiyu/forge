// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
//
//go:build linux

package unsorted

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
	"go.lindenii.runxiyu.org/forge/forged/internal/ansiec"
	"go.lindenii.runxiyu.org/forge/forged/internal/misc"
)

var (
	errGetFD    = errors.New("unable to get file descriptor")
	errGetUcred = errors.New("failed getsockopt")
)

// hooksHandler handles a connection from hookc via the
// unix socket.
func (s *Server) hooksHandler(conn net.Conn) {
	var ctx context.Context
	var cancel context.CancelFunc
	var ucred *syscall.Ucred
	var err error
	var cookie []byte
	var packPass packPass
	var sshStderr io.Writer
	var hookRet byte

	defer conn.Close()
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// There aren't reasonable cases where someone would run this as
	// another user.
	if ucred, err = getUcred(conn); err != nil {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		writeRedError(conn, "\nUnable to get peer credentials: %v", err)
		return
	}
	uint32uid := uint32(os.Getuid()) //#nosec G115
	if ucred.Uid != uint32uid {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		writeRedError(conn, "\nUID mismatch")
		return
	}

	cookie = make([]byte, 64)
	if _, err = conn.Read(cookie); err != nil {
		if _, err = conn.Write([]byte{1}); err != nil {
			return
		}
		writeRedError(conn, "\nFailed to read cookie: %v", err)
		return
	}

	{
		var ok bool
		packPass, ok = s.packPasses.Load(misc.BytesToString(cookie))
		if !ok {
			if _, err = conn.Write([]byte{1}); err != nil {
				return
			}
			writeRedError(conn, "\nInvalid handler cookie")
			return
		}
	}

	sshStderr = packPass.session.Stderr()

	_, _ = sshStderr.Write([]byte{'\n'})

	hookRet = func() byte {
		var argc64 uint64
		if err = binary.Read(conn, binary.NativeEndian, &argc64); err != nil {
			writeRedError(sshStderr, "Failed to read argc: %v", err)
			return 1
		}
		var args []string
		for range argc64 {
			var arg bytes.Buffer
			for {
				nextByte := make([]byte, 1)
				n, err := conn.Read(nextByte)
				if err != nil || n != 1 {
					writeRedError(sshStderr, "Failed to read arg: %v", err)
					return 1
				}
				if nextByte[0] == 0 {
					break
				}
				arg.WriteByte(nextByte[0])
			}
			args = append(args, arg.String())
		}

		gitEnv := make(map[string]string)
		for {
			var envLine bytes.Buffer
			for {
				nextByte := make([]byte, 1)
				n, err := conn.Read(nextByte)
				if err != nil || n != 1 {
					writeRedError(sshStderr, "Failed to read environment variable: %v", err)
					return 1
				}
				if nextByte[0] == 0 {
					break
				}
				envLine.WriteByte(nextByte[0])
			}
			if envLine.Len() == 0 {
				break
			}
			kv := envLine.String()
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) < 2 {
				writeRedError(sshStderr, "Invalid environment variable line: %v", kv)
				return 1
			}
			gitEnv[parts[0]] = parts[1]
		}

		var stdin bytes.Buffer
		if _, err = io.Copy(&stdin, conn); err != nil {
			writeRedError(conn, "Failed to read to the stdin buffer: %v", err)
		}

		switch filepath.Base(args[0]) {
		case "pre-receive":
			if packPass.directAccess {
				return 0
			}
			allOK := true
			for {
				var line, oldOID, rest, newIOID, refName string
				var found bool
				var oldHash, newHash plumbing.Hash
				var oldCommit, newCommit *object.Commit
				var pushOptCount int

				pushOptCount, err = strconv.Atoi(gitEnv["GIT_PUSH_OPTION_COUNT"])
				if err != nil {
					writeRedError(sshStderr, "Failed to parse GIT_PUSH_OPTION_COUNT: %v", err)
					return 1
				}

				// TODO: Allow existing users (even if they are already federated or registered) to add a federated user ID... though perhaps this should be in the normal SSH interface instead of the git push interface?
				// Also it'd be nice to be able to combine users or whatever
				if packPass.contribReq == "federated" && packPass.userType != "federated" && packPass.userType != "registered" {
					if pushOptCount == 0 {
						writeRedError(sshStderr, "This repo requires contributors to be either federated or registered users. You must supply your federated user ID as a push option. For example, git push -o fedid=sr.ht:runxiyu")
						return 1
					}
					for pushOptIndex := range pushOptCount {
						pushOpt, ok := gitEnv[fmt.Sprintf("GIT_PUSH_OPTION_%d", pushOptIndex)]
						if !ok {
							writeRedError(sshStderr, "Failed to get push option %d", pushOptIndex)
							return 1
						}
						if strings.HasPrefix(pushOpt, "fedid=") {
							fedUserID := strings.TrimPrefix(pushOpt, "fedid=")
							service, username, found := strings.Cut(fedUserID, ":")
							if !found {
								writeRedError(sshStderr, "Invalid federated user identifier %#v does not contain a colon", fedUserID)
								return 1
							}

							ok, err := s.fedauth(ctx, packPass.userID, service, username, packPass.pubkey)
							if err != nil {
								writeRedError(sshStderr, "Failed to verify federated user identifier %#v: %v", fedUserID, err)
								return 1
							}
							if !ok {
								writeRedError(sshStderr, "Failed to verify federated user identifier %#v: you don't seem to be on the list", fedUserID)
								return 1
							}

							break
						}
						if pushOptIndex == pushOptCount-1 {
							writeRedError(sshStderr, "This repo requires contributors to be either federated or registered users. You must supply your federated user ID as a push option. For example, git push -o fedid=sr.ht:runxiyu")
							return 1
						}
					}
				}

				line, err = stdin.ReadString('\n')
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					writeRedError(sshStderr, "Failed to read pre-receive line: %v", err)
					return 1
				}
				line = line[:len(line)-1]

				oldOID, rest, found = strings.Cut(line, " ")
				if !found {
					writeRedError(sshStderr, "Invalid pre-receive line: %v", line)
					return 1
				}

				newIOID, refName, found = strings.Cut(rest, " ")
				if !found {
					writeRedError(sshStderr, "Invalid pre-receive line: %v", line)
					return 1
				}

				if strings.HasPrefix(refName, "refs/heads/contrib/") {
					if allZero(oldOID) { // New branch
						fmt.Fprintln(sshStderr, ansiec.Blue+"POK"+ansiec.Reset, refName)
						var newMRLocalID int

						if packPass.userID != 0 {
							err = s.database.QueryRow(ctx,
								"INSERT INTO merge_requests (repo_id, creator, source_ref, status) VALUES ($1, $2, $3, 'open') RETURNING repo_local_id",
								packPass.repoID, packPass.userID, strings.TrimPrefix(refName, "refs/heads/"),
							).Scan(&newMRLocalID)
						} else {
							err = s.database.QueryRow(ctx,
								"INSERT INTO merge_requests (repo_id, source_ref, status) VALUES ($1, $2, 'open') RETURNING repo_local_id",
								packPass.repoID, strings.TrimPrefix(refName, "refs/heads/"),
							).Scan(&newMRLocalID)
						}
						if err != nil {
							writeRedError(sshStderr, "Error creating merge request: %v", err)
							return 1
						}
						mergeRequestWebURL := fmt.Sprintf("%s/contrib/%d/", s.genHTTPRemoteURL(packPass.groupPath, packPass.repoName), newMRLocalID)
						fmt.Fprintln(sshStderr, ansiec.Blue+"Created merge request at", mergeRequestWebURL+ansiec.Reset)

						s.ircBot.Send("PRIVMSG #chat :New merge request at " + mergeRequestWebURL)
					} else { // Existing contrib branch
						var existingMRUser int
						var isAncestor bool

						err = s.database.QueryRow(ctx,
							"SELECT COALESCE(creator, 0) FROM merge_requests WHERE source_ref = $1 AND repo_id = $2",
							strings.TrimPrefix(refName, "refs/heads/"), packPass.repoID,
						).Scan(&existingMRUser)
						if err != nil {
							if errors.Is(err, pgx.ErrNoRows) {
								writeRedError(sshStderr, "No existing merge request for existing contrib branch: %v", err)
							} else {
								writeRedError(sshStderr, "Error querying for existing merge request: %v", err)
							}
							return 1
						}
						if existingMRUser == 0 {
							allOK = false
							fmt.Fprintln(sshStderr, ansiec.Red+"NAK"+ansiec.Reset, refName, "(branch belongs to unowned MR)")
							continue
						}

						if existingMRUser != packPass.userID {
							allOK = false
							fmt.Fprintln(sshStderr, ansiec.Red+"NAK"+ansiec.Reset, refName, "(branch belongs another user's MR)")
							continue
						}

						oldHash = plumbing.NewHash(oldOID)

						if oldCommit, err = packPass.repo.CommitObject(oldHash); err != nil {
							writeRedError(sshStderr, "Daemon failed to get old commit: %v", err)
							return 1
						}

						// Potential BUG: I'm not sure if new_commit is guaranteed to be
						// detectable as they haven't been merged into the main repo's
						// objects yet. But it seems to work, and I don't think there's
						// any reason for this to only work intermitently.
						newHash = plumbing.NewHash(newIOID)
						if newCommit, err = packPass.repo.CommitObject(newHash); err != nil {
							writeRedError(sshStderr, "Daemon failed to get new commit: %v", err)
							return 1
						}

						if isAncestor, err = oldCommit.IsAncestor(newCommit); err != nil {
							writeRedError(sshStderr, "Daemon failed to check if old commit is ancestor: %v", err)
							return 1
						}

						if !isAncestor {
							// TODO: Create MR snapshot ref instead
							allOK = false
							fmt.Fprintln(sshStderr, ansiec.Red+"NAK"+ansiec.Reset, refName, "(force pushes are not supported yet)")
							continue
						}

						fmt.Fprintln(sshStderr, ansiec.Blue+"POK"+ansiec.Reset, refName)
					}
				} else { // Non-contrib branch
					allOK = false
					fmt.Fprintln(sshStderr, ansiec.Red+"NAK"+ansiec.Reset, refName, "(you cannot push to branches outside of contrib/*)")
				}
			}

			fmt.Fprintln(sshStderr)
			if allOK {
				fmt.Fprintln(sshStderr, "Overall "+ansiec.Green+"ACK"+ansiec.Reset+" (all checks passed)")
				return 0
			}
			fmt.Fprintln(sshStderr, "Overall "+ansiec.Red+"NAK"+ansiec.Reset+" (one or more branches failed checks)")
			return 1
		default:
			fmt.Fprintln(sshStderr, ansiec.Red+"Invalid hook:", args[0]+ansiec.Reset)
			return 1
		}
	}()

	fmt.Fprintln(sshStderr)

	_, _ = conn.Write([]byte{hookRet})
}

// serveGitHooks handles connections on the specified network listener and
// treats incoming connections as those from git hook handlers by spawning
// sessions. The listener must be a SOCK_STREAM UNIX domain socket. The
// function itself blocks.
func (s *Server) serveGitHooks(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.hooksHandler(conn)
	}
}

// getUcred fetches connection credentials as a [syscall.Ucred] from a given
// [net.Conn]. It panics when conn is not a [net.UnixConn].
func getUcred(conn net.Conn) (ucred *syscall.Ucred, err error) {
	unixConn := conn.(*net.UnixConn)
	var unixConnFD *os.File

	if unixConnFD, err = unixConn.File(); err != nil {
		return nil, errGetFD
	}
	defer unixConnFD.Close()

	if ucred, err = syscall.GetsockoptUcred(int(unixConnFD.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED); err != nil {
		return nil, errGetUcred
	}
	return ucred, nil
}

// allZero returns true if all runes in a given string are '0'. The comparison
// is not constant time and must not be used in contexts where time-based side
// channel attacks are a concern.
func allZero(s string) bool {
	for _, r := range s {
		if r != '0' {
			return false
		}
	}
	return true
}
