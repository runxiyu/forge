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
		dprintf(STDERR_FILENO, "environment variable LINDENII_FORGE_HOOKS_COOKIE is not 64 characters long, something has gone wrong\n");
		dprintf(STDERR_FILENO, "%s\n", cookie);
		return EXIT_FAILURE;
	}

	/*
	 * All hooks in git (see builtin/receive-pack.c) use a pipe by setting
	 * .in = -1 on the child_process struct, which enables us to use
	 * splice(2) to move the data to the UNIX domain socket. Just to be
	 * safe, we check that stdin is a pipe; and additionally we fetch the
	 * buffer size of the pipe to use as the maximum size for the splice.
	 *
	 * We connect to the UNIX domain socket after ensuring that standard
	 * input matches our expectations.
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
	int stdin_pipe_size = fcntl(STDIN_FILENO, F_GETPIPE_SZ);
	if (stdin_pipe_size == -1) {
		perror("fcntl on stdin");
		return EXIT_FAILURE;
	}

	/*
	 * ... And we do the same for stderr. Later we will splice from the
	 * socket to stderr, to let the daemon report back to the user.
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
	int stderr_pipe_size = fcntl(STDERR_FILENO, F_GETPIPE_SZ);
	if (stderr_pipe_size == -1) {
		perror("fcntl on stderr");
		return EXIT_FAILURE;
	}

	/*
	 * Now that we know that stdin and stderr are pipes, we can connect to
	 * the UNIX domain socket. We don't do this earlier because we don't
	 * want to create unnecessary connections if the hook was called
	 * inappropriately (such as by a user with shell access in their
	 * terminal).
	 */
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
	 * We first send the 64-byte cookie to the UNIX domain socket
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
	 * Next we can report argc and argv to the UNIX domain socket.
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
	 * Now we can start splicing data from stdin to the UNIX domain socket.
	 * The format is irrelevant and depends on the hook being called. All we
	 * do is pass it to the socket for it to handle.
	 */
	ssize_t stdin_bytes_spliced;
	while ((stdin_bytes_spliced = splice(STDIN_FILENO, NULL, sock, NULL, stdin_pipe_size, SPLICE_F_MORE)) > 0) {
	}
	if (stdin_bytes_spliced == -1) {
		perror("splice stdin to internal socket");
		close(sock);
		return EXIT_FAILURE;
	}

	/*
	 * The first byte of the response from the UNIX domain socket is the
	 * status code. We read it and record it as our return value.
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
	 */
	ssize_t stderr_bytes_spliced;
	while ((stderr_bytes_spliced = splice(sock, NULL, STDERR_FILENO, NULL, stderr_pipe_size, SPLICE_F_MORE)) > 0) {
	}
	if (stdin_bytes_spliced == -1 && errno != ECONNRESET) {
		perror("splice internal socket to stderr");
		close(sock);
		return EXIT_FAILURE;
	}

	close(sock);
	return *status_buf;
}
