// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// deploy_hooks_to_filesystem deploys the git hooks client to the filesystem.
// The git hooks client is expected to be embedded in resources_fs and must be
// pre-compiled during the build process; see the Makefile.
func deploy_hooks_to_filesystem() (err error) {
	err = func() (err error) {
		var src_fd fs.File
		var dst_fd *os.File

		if src_fd, err = resources_fs.Open("git_hooks_client/git_hooks_client"); err != nil {
			return err
		}
		defer src_fd.Close()

		if dst_fd, err = os.OpenFile(filepath.Join(config.Hooks.Execs, "git_hooks_client"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755); err != nil {
			return err
		}
		defer dst_fd.Close()

		if _, err = io.Copy(dst_fd, src_fd); err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return err
	}

	// Go's embed filesystems do not store permissions; but in any case,
	// they would need to be 0o755:
	if err = os.Chmod(filepath.Join(config.Hooks.Execs, "git_hooks_client"), 0o755); err != nil {
		return err
	}

	for _, hook_name := range []string{
		"pre-receive",
	} {
		if err = os.Symlink(filepath.Join(config.Hooks.Execs, "git_hooks_client"), filepath.Join(config.Hooks.Execs, hook_name)); err != nil {
			return err
		}
	}

	return nil
}
