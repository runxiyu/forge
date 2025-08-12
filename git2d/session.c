/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

void *session(void *_conn)
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
	char path[4096] = { 0 };
	err = bare_get_data(&reader, (uint8_t *) path, sizeof(path) - 1);
	if (err != BARE_ERROR_NONE) {
		goto close;
	}
	path[sizeof(path) - 1] = '\0';
	fprintf(stderr, "session: path='%s'\n", path);

	/* Command */
	uint64_t cmd = 0;
	err = bare_get_uint(&reader, &cmd);
	if (err != BARE_ERROR_NONE) {
		bare_put_uint(&writer, 2);
		goto close;
	}
	fprintf(stderr, "session: cmd=%llu\n", (unsigned long long)cmd);

	/* Repo init does not require opening an existing repo so let's just do it here */
	if (cmd == 15) {
		fprintf(stderr, "session: handling init for '%s'\n", path);
		if (cmd_init_repo(path, &reader, &writer) != 0) {
		}
		goto close;
	}

	git_repository *repo = NULL;
	err = git_repository_open_ext(&repo, path, GIT_REPOSITORY_OPEN_NO_SEARCH | GIT_REPOSITORY_OPEN_BARE | GIT_REPOSITORY_OPEN_NO_DOTGIT, NULL);
	if (err != 0) {
		bare_put_uint(&writer, 1);
		goto close;
	}
	switch (cmd) {
	case 1:
		err = cmd_index(repo, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 2:
		err = cmd_treeraw(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 3:
		err = cmd_resolve_ref(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 4:
		err = cmd_list_branches(repo, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 5:
		err = cmd_format_patch(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
case 6:
	err = cmd_commit_info(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 7:
		err = cmd_merge_base(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 8:
		err = cmd_log(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 9:
		err = cmd_tree_list_by_oid(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 10:
		err = cmd_write_tree(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 11:
		err = cmd_blob_write(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 12:
		err = cmd_commit_tree_oid(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 13:
		err = cmd_commit_create(repo, &reader, &writer);
		if (err != 0)
			goto free_repo;
		break;
	case 14:
		err = cmd_update_ref(repo, &reader, &writer);
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
