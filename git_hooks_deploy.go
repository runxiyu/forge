package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

func deploy_hooks_to_filesystem() (err error) {
	err = func() error {
		src_fd, err := resources_fs.Open("git_hooks_client/git_hooks_client")
		if err != nil {
			return err
		}
		defer src_fd.Close()

		dst_fd, err := os.OpenFile(filepath.Join(config.Hooks.Execs, "git_hooks_client"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		defer dst_fd.Close()

		_, err = io.Copy(dst_fd, src_fd)
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return err
	}

	err = os.Chmod(filepath.Join(config.Hooks.Execs, "git_hooks_client"), 0o755)
	if err != nil {
		return err
	}

	for _, hook_name := range []string{
		"pre-receive",
	} {
		err = os.Symlink(filepath.Join(config.Hooks.Execs, "git_hooks_client"), filepath.Join(config.Hooks.Execs, hook_name))
		if err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	return nil
}
