/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

int cmd_init_repo(const char *path, struct bare_reader *reader, struct bare_writer *writer)
{
	char hooks[4096] = { 0 };
	if (bare_get_data(reader, (uint8_t *) hooks, sizeof(hooks) - 1) != BARE_ERROR_NONE) {
		fprintf(stderr, "init_repo: protocol error reading hooks for path '%s'\n", path);
		bare_put_uint(writer, 11);
		return -1;
	}

	fprintf(stderr, "init_repo: starting for path='%s' hooks='%s'\n", path, hooks);

	if (mkdir(path, 0700) != 0 && errno != EEXIST) {
		fprintf(stderr, "init_repo: mkdir failed for '%s': %s\n", path, strerror(errno));
		bare_put_uint(writer, 24);
		return -1;
	}

	git_repository *repo = NULL;
	git_repository_init_options opts;
	git_repository_init_options_init(&opts, GIT_REPOSITORY_INIT_OPTIONS_VERSION);
	opts.flags = GIT_REPOSITORY_INIT_BARE;
	if (git_repository_init_ext(&repo, path, &opts) != 0) {
		const git_error *ge = git_error_last();
		fprintf(stderr, "init_repo: git_repository_init_ext failed: %s (klass=%d)\n", ge && ge->message ? ge->message : "(no message)", ge ? ge->klass : 0);
		bare_put_uint(writer, 20);
		return -1;
	}
	git_config *cfg = NULL;
	if (git_repository_config(&cfg, repo) != 0) {
		git_repository_free(repo);
		const git_error *ge = git_error_last();
		fprintf(stderr, "init_repo: open config failed: %s (klass=%d)\n", ge && ge->message ? ge->message : "(no message)", ge ? ge->klass : 0);
		bare_put_uint(writer, 21);
		return -1;
	}
	if (git_config_set_string(cfg, "core.hooksPath", hooks) != 0) {
		git_config_free(cfg);
		git_repository_free(repo);
		const git_error *ge = git_error_last();
		fprintf(stderr, "init_repo: set hooksPath failed: %s (klass=%d) hooks='%s'\n", ge && ge->message ? ge->message : "(no message)", ge ? ge->klass : 0, hooks);
		bare_put_uint(writer, 22);
		return -1;
	}
	if (git_config_set_bool(cfg, "receive.advertisePushOptions", 1) != 0) {
		git_config_free(cfg);
		git_repository_free(repo);
		const git_error *ge = git_error_last();
		fprintf(stderr, "init_repo: set advertisePushOptions failed: %s (klass=%d)\n", ge && ge->message ? ge->message : "(no message)", ge ? ge->klass : 0);
		bare_put_uint(writer, 23);
		return -1;
	}
	git_config_free(cfg);

	git_repository_free(repo);
	fprintf(stderr, "init_repo: success for path='%s'\n", path);
	bare_put_uint(writer, 0);
	return 0;
}
