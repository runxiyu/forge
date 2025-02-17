package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
)

func handle_upload_pack(w http.ResponseWriter, r *http.Request, params map[string]any) (err error) {
	group_name, repo_name := params["group_name"].(string), params["repo_name"].(string)
	var repo_path string
	err = database.QueryRow(r.Context(), "SELECT r.filesystem_path FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1 AND r.name = $2;", group_name, repo_name).Scan(&repo_path)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", repo_path)
	cmd.Env = append(os.Environ(), "LINDENII_FORGE_HOOKS_SOCKET_PATH="+config.Hooks.Socket)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	defer func() {
		_ = stdout.Close()
	}()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer func() {
		_ = stdin.Close()
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	_, err = io.Copy(stdin, r.Body)
	if err != nil {
		return err
	}

	err = stdin.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, stdout)
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
