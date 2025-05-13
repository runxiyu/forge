/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

/*
 * TODO: Pool repositories (and take care of thread safety)
 * libgit2 has a nice builtin per-repo cache that we could utilize this way.
 */

#include "x.h"

int
main(int argc, char **argv)
{
	if (argc != 2) {
		errx(1, "provide one argument: the socket path");
	}

	signal(SIGPIPE, SIG_IGN);

	git_libgit2_init();

	int sock;
	if ((sock = socket(AF_UNIX, SOCK_STREAM | SOCK_CLOEXEC, 0)) < 0)
		err(1, "socket");

	struct sockaddr_un addr;
	memset(&addr, 0, sizeof(addr));
	addr.sun_family = AF_UNIX;
	strcpy(addr.sun_path, argv[1]);

	umask(0077);

	if (bind(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un))) {
		if (errno == EADDRINUSE) {
			unlink(argv[1]);
			if (bind(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un)))
				err(1, "bind");
		} else {
			err(1, "bind");
		}
	}

	listen(sock, 128);

	for (;;) {
		int *conn = malloc(sizeof(int));
		if (conn == NULL)
			err(1, "malloc");

		*conn = accept(sock, 0, 0);
		if (*conn == -1)
			err(1, "accept");

		pthread_t thread;

		pthread_create(&thread, NULL, session, (void *)conn);
	}

	close(sock);

	git_libgit2_shutdown();
}
