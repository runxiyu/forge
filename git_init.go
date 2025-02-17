package main

import (
	"github.com/go-git/go-git/v5"
	git_format_config "github.com/go-git/go-git/v5/plumbing/format/config"
)

func git_bare_init_with_default_hooks(repo_path string) (err error) {
	repo, err := git.PlainInit(repo_path, true)
	if err != nil {
		return err
	}

	git_config, err := repo.Config()
	if err != nil {
		return err
	}

	git_config.Raw.SetOption("core", git_format_config.NoSubsection, "hooksPath", config.Hooks.Execs)

	err = repo.SetConfig(git_config)
	if err != nil {
		return err
	}

	return nil
}
