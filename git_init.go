// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package main

import (
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	gitFmtConfig "github.com/go-git/go-git/v5/plumbing/format/config"
)

// gitInit initializes a bare git repository with the forge-deployed hooks
// directory as the hooksPath.
func (s *server) gitInit(repoPath string) (err error) {
	var repo *git.Repository
	var gitConf *gitConfig.Config

	if repo, err = git.PlainInit(repoPath, true); err != nil {
		return err
	}

	if gitConf, err = repo.Config(); err != nil {
		return err
	}

	gitConf.Raw.SetOption("core", gitFmtConfig.NoSubsection, "hooksPath", s.config.Hooks.Execs)
	gitConf.Raw.SetOption("receive", gitFmtConfig.NoSubsection, "advertisePushOptions", "true")

	if err = repo.SetConfig(gitConf); err != nil {
		return err
	}

	return nil
}
