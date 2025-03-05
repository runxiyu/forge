// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
)

func handle_repo_info(w http.ResponseWriter, r *http.Request, params map[string]any) (err error) {
	var group_name, repo_name, repo_path string

	group_name, repo_name = params["group_name"].(string), params["repo_name"].(string)
	if err = database.QueryRow(r.Context(),
		"SELECT r.filesystem_path FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1 AND r.name = $2;",
		group_name, repo_name,
	).Scan(&repo_path); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	w.WriteHeader(http.StatusOK)

	cmd := exec.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", repo_path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer func() {
		_ = stdout.Close()
	}()
	cmd.Stderr = cmd.Stdout

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = pack_line(w, "# service=git-upload-pack\n"); err != nil {
		return err
	}

	if err = pack_flush(w); err != nil {
		return
	}

	if _, err = io.Copy(w, stdout); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// Taken from https://github.com/icyphox/legit, MIT license
func pack_line(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "%04x%s", len(s)+4, s)
	return err
}

// Taken from https://github.com/icyphox/legit, MIT license
func pack_flush(w io.Writer) error {
	_, err := fmt.Fprint(w, "0000")
	return err
}
