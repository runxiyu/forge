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
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "bare.h"

typedef struct {
	int fd;
} conn_io_t;

static bare_error
conn_read(void *buffer, void *dst, uint64_t sz)
{
	conn_io_t *io = buffer;
	ssize_t rsz = read(io->fd, dst, sz);
	return (rsz == (ssize_t)sz) ? BARE_ERROR_NONE : BARE_ERROR_READ_FAILED;
}

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

	char path[4096];
	conn_io_t io = {.fd = conn};
	struct bare_reader reader = {
		.buffer = &io,
		.read = conn_read,
	};

	err = bare_get_data(&reader, (uint8_t *)path, sizeof(path) - 1);
	if (err != BARE_ERROR_NONE) {
		ret = 10;
		goto close;
	}
	path[sizeof(path) - 1] = '\0';

	err = git_repository_open_ext(&repo, path, GIT_REPOSITORY_OPEN_NO_SEARCH | GIT_REPOSITORY_OPEN_BARE | GIT_REPOSITORY_OPEN_NO_DOTGIT, NULL);
	if (err != 0) {
		ret = 1;
		goto close;
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
main(int argc, char **argv)
{
	if (argc != 2) {
		errx(1, "provide one argument: the socket path");
	}

	git_libgit2_init();

	int		sock;
	if ((sock = socket(AF_UNIX, SOCK_STREAM | SOCK_CLOEXEC, 0)) < 0)
		err(1, "socket");

	struct sockaddr_un addr;
	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strcpy(addr.sun_path, argv[1]);

	if (bind(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un))) {
		if (errno == EADDRINUSE) {
			unlink(argv[1]);
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

		pthread_t	thread;

		pthread_create(&thread, NULL, session, (void *)conn);
	}

	close(sock);

	git_libgit2_shutdown();
}

