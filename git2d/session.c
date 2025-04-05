/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

void *
session(void *_conn)
{
	int conn = *(int *)_conn;
	free((int *)_conn);

	int err;

	conn_io_t io = {.fd = conn };
	struct bare_reader reader = {
		.buffer = &io,
		.read = conn_read,
	};
	struct bare_writer writer = {
		.buffer = &io,
		.write = conn_write,
	};

	/* Repo path */
	char path[4096] = {0};
	err = bare_get_data(&reader, (uint8_t *) path, sizeof(path) - 1);
	if (err != BARE_ERROR_NONE) {
		goto close;
	}
	path[sizeof(path) - 1] = '\0';

	/* Open repo */
	git_repository *repo = NULL;
	err = git_repository_open_ext(&repo, path, GIT_REPOSITORY_OPEN_NO_SEARCH | GIT_REPOSITORY_OPEN_BARE | GIT_REPOSITORY_OPEN_NO_DOTGIT, NULL);
	if (err != 0) {
		bare_put_uint(&writer, 1);
		goto close;
	}

	/* Command */
	uint64_t cmd = 0;
	err = bare_get_uint(&reader, &cmd);
	if (err != BARE_ERROR_NONE) {
		bare_put_uint(&writer, 2);
		goto free_repo;
	}
	switch (cmd) {
	case 1:
		err = cmd1(repo, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 0:
		bare_put_uint(&writer, 3);
		goto free_repo;
	default:
		bare_put_uint(&writer, 3);
		goto free_repo;
	}

free_repo:
	git_repository_free(repo);

close:
	close(conn);

	return NULL;
}

