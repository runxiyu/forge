// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

package main

import (
	"github.com/go-git/go-git/v5"
	git_config "github.com/go-git/go-git/v5/config"
	git_format_config "github.com/go-git/go-git/v5/plumbing/format/config"
)

// git_bare_init_with_default_hooks initializes a bare git repository with the
// forge-deployed hooks directory as the hooksPath.
func git_bare_init_with_default_hooks(repo_path string) (err error) {
	var repo *git.Repository
	var git_config *git_config.Config

	if repo, err = git.PlainInit(repo_path, true); err != nil {
		return err
	}

	if git_config, err = repo.Config(); err != nil {
		return err
	}

	git_config.Raw.SetOption("core", git_format_config.NoSubsection, "hooksPath", config.Hooks.Execs)

	if err = repo.SetConfig(git_config); err != nil {
		return err
	}

	return nil
}
