package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	glider_ssh "github.com/gliderlabs/ssh"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
)

type pack_to_hook_t struct {
	session       glider_ssh.Session
	pubkey        string
	direct_access bool
	repo_path     string
	user_id       int
	repo_id       int
	group_name    string
	repo_name     string
}

var pack_to_hook_by_cookie = cmap.Map[string, pack_to_hook_t]{}

// ssh_handle_receive_pack handles attempts to push to repos.
func ssh_handle_receive_pack(session glider_ssh.Session, pubkey string, repo_identifier string) (err error) {
	group_name, repo_name, repo_id, repo_path, direct_access, contrib_requirements, user_type, user_id, err := get_repo_path_perms_from_ssh_path_pubkey(session.Context(), repo_identifier, pubkey)
	if err != nil {
		return err
	}

	if !direct_access {
		switch contrib_requirements {
		case "closed":
			if !direct_access {
				return errors.New("You need direct access to push to this repo.")
			}
		case "registered_user":
			if user_type != "registered" {
				return errors.New("You need to be a registered user to push to this repo.")
			}
		case "ssh_pubkey":
			if pubkey == "" {
				return errors.New("You need to have an SSH public key to push to this repo.")
			}
			if user_type == "" {
				user_id, err = add_user_ssh(session.Context(), pubkey)
				if err != nil {
					return err
				}
				fmt.Fprintln(session.Stderr(), "You are now registered as user ID", user_id)
			}
		case "public":
		default:
			panic("unknown contrib_requirements value " + contrib_requirements)
		}
	}

	cookie, err := random_urlsafe_string(16)
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while generating cookie:", err)
	}

	fmt.Println(group_name, repo_name)

	pack_to_hook_by_cookie.Store(cookie, pack_to_hook_t{
		session:       session,
		pubkey:        pubkey,
		direct_access: direct_access,
		repo_path:     repo_path,
		user_id:       user_id,
		repo_id:       repo_id,
		group_name:    group_name,
		repo_name:     repo_name,
	})
	defer pack_to_hook_by_cookie.Delete(cookie)
	// The Delete won't execute until proc.Wait returns unless something
	// horribly wrong such as a panic occurs.

	proc := exec.CommandContext(session.Context(), "git-receive-pack", repo_path)
	proc.Env = append(os.Environ(),
		"LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket,
		"LINDENII_FORGE_HOOKS_COOKIE="+cookie,
	)
	proc.Stdin = session
	proc.Stdout = session
	proc.Stderr = session.Stderr()

	err = proc.Start()
	if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while starting process:", err)
		return err
	}

	err = proc.Wait()
	if exitError, ok := err.(*exec.ExitError); ok {
		fmt.Fprintln(session.Stderr(), "Process exited with error", exitError.ExitCode())
	} else if err != nil {
		fmt.Fprintln(session.Stderr(), "Error while waiting for process:", err)
	}

	return err
}
