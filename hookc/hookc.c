/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileContributor: Runxi Yu <https://runxiyu.org>
 * SPDX-FileContributor: Test_User <hax@runxiyu.org>
 */

#include <errno.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/stat.h>
#include <string.h>
#include <fcntl.h>
#include <signal.h>
#ifdef __linux__
#include <linux/limits.h>
#include <sys/sendfile.h>
#define USE_SPLICE 1
#else
#define USE_SPLICE 0
#endif

int main(int argc, char *argv[]) {
	if (signal(SIGPIPE, SIG_IGN) == SIG_ERR) {
		perror("signal");
		return EXIT_FAILURE;
	}

	const char *socket_path = getenv("LINDENII_FORGE_HOOKS_SOCKET_PATH");
	if (socket_path == NULL) {
		dprintf(STDERR_FILENO, "environment variable LINDENII_FORGE_HOOKS_SOCKET_PATH undefined\n");
		return EXIT_FAILURE;
	}
	const char *cookie = getenv("LINDENII_FORGE_HOOKS_COOKIE");
	if (cookie == NULL) {
		dprintf(STDERR_FILENO, "environment variable LINDENII_FORGE_HOOKS_COOKIE undefined\n");
		return EXIT_FAILURE;
	}
	if (strlen(cookie) != 64) {
		dprintf(STDERR_FILENO, "environment variable LINDENII_FORGE_HOOKS_COOKIE is not 64 characters long\n");
		return EXIT_FAILURE;
	}

	/*
	 * All hooks in git (see builtin/receive-pack.c) use a pipe by setting
	 * .in = -1 on the child_process struct, which enables us to use
	 * splice(2) to move the data to the UNIX domain socket.
	 */
	struct stat stdin_stat;
	if (fstat(STDIN_FILENO, &stdin_stat) == -1) {
		perror("fstat on stdin");
		return EXIT_FAILURE;
	}
	if (!S_ISFIFO(stdin_stat.st_mode)) {
		dprintf(STDERR_FILENO, "stdin must be a pipe\n");
		return EXIT_FAILURE;
	}
	#if USE_SPLICE
	int stdin_pipe_size = fcntl(STDIN_FILENO, F_GETPIPE_SZ);
	if (stdin_pipe_size == -1) {
		perror("fcntl on stdin");
		return EXIT_FAILURE;
	}
#else
	int stdin_pipe_size = 65536;
#endif

	if (stdin_pipe_size == -1) {
		perror("fcntl on stdin");
		return EXIT_FAILURE;
	}

	/*
	 * Same for stderr.
	 */
	struct stat stderr_stat;
	if (fstat(STDERR_FILENO, &stderr_stat) == -1) {
		perror("fstat on stderr");
		return EXIT_FAILURE;
	}
	if (!S_ISFIFO(stderr_stat.st_mode)) {
		dprintf(STDERR_FILENO, "stderr must be a pipe\n");
		return EXIT_FAILURE;
	}
#if USE_SPLICE
	int stderr_pipe_size = fcntl(STDERR_FILENO, F_GETPIPE_SZ);
	if (stderr_pipe_size == -1) {
		perror("fcntl on stderr");
		return EXIT_FAILURE;
	}
#else
	int stderr_pipe_size = 65536;
#endif
	if (stderr_pipe_size == -1) {
		perror("fcntl on stderr");
		return EXIT_FAILURE;
	}

	/* Connecting back to the main daemon */
	int sock;
	struct sockaddr_un addr;
	sock = socket(AF_UNIX, SOCK_STREAM, 0);
	if (sock == -1) {
		perror("internal socket creation");
		return EXIT_FAILURE;
	}
	memset(&addr, 0, sizeof(struct sockaddr_un));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path) - 1);
	if (connect(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un)) == -1) {
		perror("internal socket connect");
		close(sock);
		return EXIT_FAILURE;
	}

	/*
	 * Send the 64-byte cookit back.
	 */
	ssize_t cookie_bytes_sent = send(sock, cookie, 64, 0);
	switch (cookie_bytes_sent) {
	case -1:
		perror("send cookie");
		close(sock);
		return EXIT_FAILURE;
	case 64:
		break;
	default:
		dprintf(STDERR_FILENO, "send returned unexpected value on internal socket\n");
		close(sock);
		return EXIT_FAILURE;
	}

	/*
	 * Report arguments.
	 */
	uint64_t argc64 = (uint64_t)argc;
	ssize_t bytes_sent = send(sock, &argc64, sizeof(argc64), 0);
	switch (bytes_sent) {
	case -1:
		perror("send argc");
		close(sock);
		return EXIT_FAILURE;
	case sizeof(argc64):
		break;
	default:
		dprintf(STDERR_FILENO, "send returned unexpected value on internal socket\n");
		close(sock);
		return EXIT_FAILURE;
	}
	for (int i = 0; i < argc; i++) {
		unsigned long len = strlen(argv[i]) + 1;
		bytes_sent = send(sock, argv[i], len, 0);
		if (bytes_sent == -1) {
			perror("send argv");
			close(sock);
			exit(EXIT_FAILURE);
		} else if ((unsigned long)bytes_sent == len) {
		} else {
			dprintf(STDERR_FILENO, "send returned unexpected value on internal socket\n");
			close(sock);
			exit(EXIT_FAILURE);
		}
	}

	/*
	 * Report GIT_* environment.
	 */
	extern char **environ;
	for (char **env = environ; *env != NULL; env++) {
		if (strncmp(*env, "GIT_", 4) == 0) {
			unsigned long len = strlen(*env) + 1;
			bytes_sent = send(sock, *env, len, 0);
			if (bytes_sent == -1) {
				perror("send env");
				close(sock);
				exit(EXIT_FAILURE);
			} else if ((unsigned long)bytes_sent == len) {
			} else {
				dprintf(STDERR_FILENO, "send returned unexpected value on internal socket\n");
				close(sock);
				exit(EXIT_FAILURE);
			}
		}
	}
	bytes_sent = send(sock, "", 1, 0);
	if (bytes_sent == -1) {
		perror("send env terminator");
		close(sock);
		exit(EXIT_FAILURE);
	} else if (bytes_sent == 1) {
	} else {
		dprintf(STDERR_FILENO, "send returned unexpected value on internal socket\n");
		close(sock);
		exit(EXIT_FAILURE);
	}

	/*
	 * Splice stdin to the daemon. For pre-receive it's just old/new/ref.
	 */
	ssize_t stdin_bytes_spliced;
#if USE_SPLICE
	while ((stdin_bytes_spliced = splice(STDIN_FILENO, NULL, sock, NULL, stdin_pipe_size, SPLICE_F_MORE)) > 0) {
	}
	if (stdin_bytes_spliced == -1) {
		perror("splice stdin to internal socket");
		close(sock);
		return EXIT_FAILURE;
	}
#else
	char buf[65536];
	ssize_t n;
	while ((n = read(STDIN_FILENO, buf, sizeof(buf))) > 0) {
		if (write(sock, buf, n) != n) {
			perror("write to internal socket");
			close(sock);
			return EXIT_FAILURE;
		}
	}
	if (n < 0) {
		perror("read from stdin");
		close(sock);
		return EXIT_FAILURE;
	}
#endif

	/*
	 * The sending part of the UNIX socket should be shut down, to let
	 * io.Copy on the Go side return.
	 */
	if (shutdown(sock, SHUT_WR) == -1) {
		perror("shutdown internal socket");
		close(sock);
		return EXIT_FAILURE;
	}

	/*
	 * The first byte of the response from the UNIX domain socket is the
	 * status code to return.
	 *
	 * FIXME: It doesn't make sense to require the return value to be
	 * sent before the log message. However, if we were to keep splicing,
	 * it's difficult to get the last byte before EOF. Perhaps we could
	 * hack together some sort of OOB message or ancillary data, or perhaps
	 * even use signals.
	 */
	char status_buf[1];
	ssize_t bytes_read = read(sock, status_buf, 1);
	switch (bytes_read) {
	case -1:
		perror("read status code from internal socket");
		close(sock);
		return EXIT_FAILURE;
	case 0:
		dprintf(STDERR_FILENO, "unexpected EOF on internal socket\n");
		close(sock);
		return EXIT_FAILURE;
	case 1:
		break;
	default:
		dprintf(STDERR_FILENO, "read returned unexpected value on internal socket\n");
		close(sock);
		return EXIT_FAILURE;
	}

	/*
	 * Now we can splice data from the UNIX domain socket to stderr.
	 * This data is directly passed to the user (with "remote: " prepended).
	 *
	 * We usually don't actually use this as the daemon could easily write
	 * to the SSH connection's stderr directly anyway.
	 */
	ssize_t stderr_bytes_spliced;
#if USE_SPLICE
	while ((stderr_bytes_spliced = splice(sock, NULL, STDERR_FILENO, NULL, stderr_pipe_size, SPLICE_F_MORE)) > 0) {
	}
	if (stderr_bytes_spliced == -1 && errno != ECONNRESET) {
		perror("splice internal socket to stderr");
		close(sock);
		return EXIT_FAILURE;
	}
#else
	while ((n = read(sock, buf, sizeof(buf))) > 0) {
		if (write(STDERR_FILENO, buf, n) != n) {
			perror("write to stderr");
			close(sock);
			return EXIT_FAILURE;
		}
	}
	if (n < 0 && errno != ECONNRESET) {
		perror("read from internal socket");
		close(sock);
		return EXIT_FAILURE;
	}
#endif

	close(sock);
	return *status_buf;
}
