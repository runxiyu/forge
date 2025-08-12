/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

/*
 * TODO: Pool repositories (and take care of thread safety)
 * libgit2 has a nice builtin per-repo cache that we could utilize this way.
 */

#include "x.h"

int main(int argc, char **argv)
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

	pthread_attr_t pthread_attr;

	if (pthread_attr_init(&pthread_attr) != 0)
		err(1, "pthread_attr_init");

	if (pthread_attr_setdetachstate(&pthread_attr, PTHREAD_CREATE_DETACHED)
	    != 0)
		err(1, "pthread_attr_setdetachstate");

	for (;;) {
		int *conn = malloc(sizeof(int));
		if (conn == NULL) {
			warn("malloc");
			continue;
		}

		*conn = accept(sock, 0, 0);
		if (*conn == -1) {
			free(conn);
			warn("accept");
			continue;
		}

		pthread_t thread;

		if (pthread_create(&thread, &pthread_attr, session, (void *)conn) != 0) {
			close(*conn);
			free(conn);
			warn("pthread_create");
		}
	}

	close(sock);

	git_libgit2_shutdown();
}
