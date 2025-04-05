// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package forge

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// deployHooks deploys the git hooks client to the filesystem. The git hooks
// client is expected to be embedded in resourcesFS and must be pre-compiled
// during the build process; see the Makefile.
func (s *Server) deployHooks() (err error) {
	err = func() (err error) {
		var srcFD fs.File
		var dstFD *os.File

		if srcFD, err = embeddedResourcesFS.Open("hookc/hookc"); err != nil {
			return err
		}
		defer srcFD.Close()

		if dstFD, err = os.OpenFile(filepath.Join(s.Config.Hooks.Execs, "hookc"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755); err != nil {
			return err
		}
		defer dstFD.Close()

		if _, err = io.Copy(dstFD, srcFD); err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return err
	}

	// Go's embed filesystems do not store permissions; but in any case,
	// they would need to be 0o755:
	if err = os.Chmod(filepath.Join(s.Config.Hooks.Execs, "hookc"), 0o755); err != nil {
		return err
	}

	for _, hookName := range []string{
		"pre-receive",
	} {
		if err = os.Symlink(filepath.Join(s.Config.Hooks.Execs, "hookc"), filepath.Join(s.Config.Hooks.Execs, hookName)); err != nil {
			if !errors.Is(err, fs.ErrExist) {
				return err
			}
			// TODO: Maybe check if it points to the right place?
		}
	}

	return nil
}
