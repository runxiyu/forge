#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/stat.h>
#include <string.h>
#include <fcntl.h>

int main(void) {
	const char *socket_path = getenv("LINDENII_FORGE_HOOKS_SOCKET_PATH");
	if (socket_path == NULL) {
	        dprintf(STDERR_FILENO, "environment variable LINDENII_FORGE_HOOKS_SOCKET_PATH undefined\n");
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
	int pipe_size = fcntl(STDIN_FILENO, F_GETPIPE_SZ);
	if (pipe_size == -1) {
		perror("fcntl on stdin");
		return EXIT_FAILURE;
	}

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


	ssize_t bytes_spliced;
	while ((bytes_spliced = splice(STDIN_FILENO, NULL, sock, NULL, pipe_size, SPLICE_F_MORE)) > 0) {
	}

	if (bytes_spliced == -1) {
		perror("splice stdin to internal socket");
		close(sock);
		return EXIT_FAILURE;
	}

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

	close(sock);
	return *status_buf;
}
