/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include <err.h>
#include <errno.h>
#include <git2.h>
#include <pthread.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/un.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

void *
session(void *_conn)
{
	int		conn = *(int *)_conn;

#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wunused-but-set-variable"
	int		ret = 0;
#pragma GCC diagnostic pop

	int		err;
	git_repository *repo = NULL;

	git_libgit2_init();

	err = git_repository_open_ext(&repo, "/home/runxiyu/Lindenii/forge/test.git", GIT_REPOSITORY_OPEN_NO_SEARCH | GIT_REPOSITORY_OPEN_BARE | GIT_REPOSITORY_OPEN_NO_DOTGIT, NULL);
	if (err != 0) {
		ret = 1;
		goto free_libgit2;
	}

	git_object     *obj = NULL;
	err = git_revparse_single(&obj, repo, "HEAD^{tree}");
	if (err != 0) {
		ret = 2;
		goto free_repo;
	}
	git_tree       *tree = (git_tree *) obj;

	git_tree_entry *entry = NULL;
	err = git_tree_entry_bypath(&entry, tree, "README.md");
	if (err != 0) {
		ret = 3;
		goto free_tree;
	}

	git_otype	objtype = git_tree_entry_type(entry);
	if (objtype != GIT_OBJECT_BLOB) {
		ret = 4;
		goto free_tree_entry;
	}

	git_object     *obj2 = NULL;
	err = git_tree_entry_to_object(&obj2, repo, entry);
	if (err != 0) {
		ret = 5;
		goto free_tree_entry;
	}

	git_blob       *blob = (git_blob *) obj2;
	const void     *content = git_blob_rawcontent(blob);
	if (content == NULL) {
		ret = 6;
		goto free_blob;
	}
	write(conn, content, git_blob_rawsize(blob));

free_blob:
	git_blob_free(blob);
free_tree_entry:
	git_tree_entry_free(entry);
free_tree:
	git_tree_free(tree);
free_repo:
	git_repository_free(repo);
free_libgit2:
	git_libgit2_shutdown();

close:
	close(conn);
	free((int *)_conn);

	/* TODO: Handle ret */

	return NULL;

	/* TODO: Actually use it properly */
	if (0)
		goto close;
}

int
main(void)
{
	int		sock;
	if ((sock = socket(AF_UNIX, SOCK_STREAM | SOCK_CLOEXEC, 0)) < 0)
		err(1, "socket");

	struct sockaddr_un addr;
	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strcpy(addr.sun_path, "/home/runxiyu/Lindenii/forge/git2d.sock");

	if (bind(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un))) {
		if (errno == EADDRINUSE) {
			unlink("/home/runxiyu/Lindenii/forge/git2d.sock");
			if (bind(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un)))
				err(1, "bind");
		} else {
			err(1, "bind");
		}
	}

	listen(sock, 0);

	for (;;) {
		int		*conn = malloc(sizeof(int));
		if (conn == NULL)
			err(1, "malloc");

		*conn = accept(sock, 0, 0);
		if (*conn == -1)
			err(1, "accept");

		puts("got");

		pthread_t	thread;

		pthread_create(&thread, NULL, session, (void *)conn);
	}

	close(sock);
}
