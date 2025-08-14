// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

package gogit

import (
	"github.com/go-git/go-git/v5"
	gitFmtConfig "github.com/go-git/go-git/v5/plumbing/format/config"
)

// InitBare initializes a bare git repository at repoPath and sets the hooks path.
func InitBare(repoPath, hooksPath string) error {
	repo, err := git.PlainInit(repoPath, true)
	if err != nil {
		return err
	}

	conf, err := repo.Config()
	if err != nil {
		return err
	}

	conf.Raw.SetOption("core", gitFmtConfig.NoSubsection, "hooksPath", hooksPath)
	conf.Raw.SetOption("receive", gitFmtConfig.NoSubsection, "advertisePushOptions", "true")

	return repo.SetConfig(conf)
}
