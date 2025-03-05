// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
)

func handle_upload_pack(w http.ResponseWriter, r *http.Request, params map[string]any) (err error) {
	var group_name, repo_name string
	var repo_path string
	var stdout io.ReadCloser
	var stdin io.WriteCloser
	var cmd *exec.Cmd

	group_name, repo_name = params["group_name"].(string), params["repo_name"].(string)

	if err = database.QueryRow(r.Context(),
		"SELECT r.filesystem_path FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1 AND r.name = $2;",
		group_name, repo_name,
	).Scan(&repo_path); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	cmd = exec.Command("git", "upload-pack", "--stateless-rpc", repo_path)
	cmd.Env = append(os.Environ(), "LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket)
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	defer func() {
		_ = stdout.Close()
	}()

	if stdin, err = cmd.StdinPipe(); err != nil {
		return err
	}
	defer func() {
		_ = stdin.Close()
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	if _, err = io.Copy(stdin, r.Body); err != nil {
		return err
	}

	if err = stdin.Close(); err != nil {
		return err
	}

	if _, err = io.Copy(w, stdout); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}
