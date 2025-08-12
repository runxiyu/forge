/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

int cmd_index(git_repository *repo, struct bare_writer *writer)
{
	/* HEAD tree */

	git_object *obj = NULL;
	int err = git_revparse_single(&obj, repo, "HEAD^{tree}");
	if (err != 0) {
		bare_put_uint(writer, 4);
		return -1;
	}
	git_tree *tree = (git_tree *) obj;

	/* README */

	git_tree_entry *entry = NULL;
	err = git_tree_entry_bypath(&entry, tree, "README.md");
	if (err != 0) {
		bare_put_uint(writer, 5);
		git_tree_free(tree);
		return -1;
	}
	git_otype objtype = git_tree_entry_type(entry);
	if (objtype != GIT_OBJECT_BLOB) {
		bare_put_uint(writer, 6);
		git_tree_entry_free(entry);
		git_tree_free(tree);
		return -1;
	}
	git_object *obj2 = NULL;
	err = git_tree_entry_to_object(&obj2, repo, entry);
	if (err != 0) {
		bare_put_uint(writer, 7);
		git_tree_entry_free(entry);
		git_tree_free(tree);
		return -1;
	}
	git_blob *blob = (git_blob *) obj2;
	const void *content = git_blob_rawcontent(blob);
	if (content == NULL) {
		bare_put_uint(writer, 8);
		git_blob_free(blob);
		git_tree_entry_free(entry);
		git_tree_free(tree);
		return -1;
	}
	bare_put_uint(writer, 0);
	bare_put_data(writer, content, git_blob_rawsize(blob));

	/* Commits */

	/* TODO BUG: This might be a different commit from the displayed README due to races */

	git_revwalk *walker = NULL;
	if (git_revwalk_new(&walker, repo) != 0) {
		bare_put_uint(writer, 9);
		git_blob_free(blob);
		git_tree_entry_free(entry);
		git_tree_free(tree);
		return -1;
	}

	if (git_revwalk_push_head(walker) != 0) {
		bare_put_uint(writer, 9);
		git_revwalk_free(walker);
		git_blob_free(blob);
		git_tree_entry_free(entry);
		git_tree_free(tree);
		return -1;
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
		bare_put_data(writer, oid.id, GIT_OID_RAWSZ);

		/* Title */
		size_t msg_len = msg ? strlen(msg) : 0;
		bare_put_data(writer, (const uint8_t *)(msg ? msg : ""), msg_len);

		/* Author's name */
		const char *author_name = author ? author->name : "";
		bare_put_data(writer, (const uint8_t *)author_name, strlen(author_name));

		/* Author's email */
		const char *author_email = author ? author->email : "";
		bare_put_data(writer, (const uint8_t *)author_email, strlen(author_email));

		/* Author's date */
		/* TODO: Pass the integer instead of a string */
		time_t time = git_commit_time(commit);
		char timebuf[64];
		struct tm *tm = localtime(&time);
		if (tm)
			strftime(timebuf, sizeof(timebuf), "%Y-%m-%d %H:%M:%S", tm);
		else
			strcpy(timebuf, "unknown");
		bare_put_data(writer, (const uint8_t *)timebuf, strlen(timebuf));

		git_commit_free(commit);
		count++;
	}

	git_revwalk_free(walker);
	git_blob_free(blob);
	git_tree_entry_free(entry);
	git_tree_free(tree);

	return 0;
}
