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
	git_repository *repo = NULL;

	char path[4096] = {0};
	conn_io_t io = {.fd = conn };
	struct bare_reader reader = {
		.buffer = &io,
		.read = conn_read,
	};
	struct bare_writer writer = {
		.buffer = &io,
		.write = conn_write,
	};

	err = bare_get_data(&reader, (uint8_t *) path, sizeof(path) - 1);
	if (err != BARE_ERROR_NONE) {
		goto close;
	}
	path[sizeof(path) - 1] = '\0';

	err = git_repository_open_ext(&repo, path, GIT_REPOSITORY_OPEN_NO_SEARCH | GIT_REPOSITORY_OPEN_BARE | GIT_REPOSITORY_OPEN_NO_DOTGIT, NULL);
	if (err != 0) {
		bare_put_uint(&writer, 1);
		goto close;
	}

	uint64_t cmd = 0;
	err = bare_get_uint(&reader, &cmd);
	if (err != BARE_ERROR_NONE) {
		bare_put_uint(&writer, 2);
		goto free_repo;
	}
	switch (cmd) {
	case 1:
		break;
	case 0:
		bare_put_uint(&writer, 3);
		goto free_repo;
	default:
		bare_put_uint(&writer, 3);
		goto free_repo;
	}

	git_object *obj = NULL;
	err = git_revparse_single(&obj, repo, "HEAD^{tree}");
	if (err != 0) {
		bare_put_uint(&writer, 4);
		goto free_repo;
	}
	git_tree *tree = (git_tree *) obj;

	/* README */

	git_tree_entry *entry = NULL;
	err = git_tree_entry_bypath(&entry, tree, "README.md");
	if (err != 0) {
		bare_put_uint(&writer, 5);
		goto free_tree;
	}

	git_otype objtype = git_tree_entry_type(entry);
	if (objtype != GIT_OBJECT_BLOB) {
		bare_put_uint(&writer, 6);
		goto free_tree_entry;
	}

	git_object *obj2 = NULL;
	err = git_tree_entry_to_object(&obj2, repo, entry);
	if (err != 0) {
		bare_put_uint(&writer, 7);
		goto free_tree_entry;
	}

	git_blob *blob = (git_blob *) obj2;
	const void *content = git_blob_rawcontent(blob);
	if (content == NULL) {
		bare_put_uint(&writer, 8);
		goto free_blob;
	}

	bare_put_uint(&writer, 0);
	bare_put_data(&writer, content, git_blob_rawsize(blob));

	/* Commits */

	git_revwalk *walker = NULL;
	if (git_revwalk_new(&walker, repo) != 0) {
		bare_put_uint(&writer, 9);
		goto free_blob;
	}

	if (git_revwalk_push_head(walker) != 0) {
		bare_put_uint(&writer, 9);
		goto free_blob;
	}

	int count = 0;
	git_oid oid;
	while (count < 3 && git_revwalk_next(&oid, walker) == 0) {
		git_commit *commit = NULL;
		if (git_commit_lookup(&commit, repo, &oid) != 0)
			break;

		const char *msg = git_commit_summary(commit);
		const git_signature *author = git_commit_author(commit);

		/* ID */
		bare_put_data(&writer, oid.id, GIT_OID_RAWSZ);

		/* Title */
		size_t msg_len = msg ? strlen(msg) : 0;
		bare_put_data(&writer, (const uint8_t *)(msg ? msg : ""), msg_len);

		/* Author's name */
		const char *author_name = author ? author->name : "";
		bare_put_data(&writer, (const uint8_t *)author_name, strlen(author_name));

		/* Author's email */
		const char *author_email = author ? author->email : "";
		bare_put_data(&writer, (const uint8_t *)author_email, strlen(author_email));

		/* Author's date */
		/* TODO: Pass the integer instead of a string */
		time_t time = git_commit_time(commit);
		char timebuf[64];
		struct tm *tm = localtime(&time);
		if (tm)
			strftime(timebuf, sizeof(timebuf), "%Y-%m-%d %H:%M:%S", tm);
		else
			strcpy(timebuf, "unknown");
		bare_put_data(&writer, (const uint8_t *)timebuf, strlen(timebuf));

		git_commit_free(commit);
		count++;
	}

	git_revwalk_free(walker);

free_blob:
	git_blob_free(blob);
free_tree_entry:
	git_tree_entry_free(entry);
free_tree:
	git_tree_free(tree);
free_repo:
	git_repository_free(repo);

close:
	close(conn);

	return NULL;
}

