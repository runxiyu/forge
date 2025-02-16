#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/stat.h>
#include <string.h>
#include <fcntl.h>

int main(void) {
	int sock;
	struct sockaddr_un addr;
	const char *socket_path = getenv("LINDENII_FORGE_HOOKS_SOCKET_PATH");

	if (socket_path == NULL) {
	        dprintf(STDERR_FILENO, "fatal: environment variable LINDENII_FORGE_HOOKS_SOCKET_PATH undefined\n");
		return EXIT_FAILURE;
	}

	sock = socket(AF_UNIX, SOCK_STREAM, 0);
	if (sock == -1) {
		perror("socket");
		return EXIT_FAILURE;
	}

	memset(&addr, 0, sizeof(struct sockaddr_un));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path) - 1);

	if (connect(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un)) == -1) {
		perror("connect");
		close(sock);
		return EXIT_FAILURE;
	}

	struct stat stdin_stat;
	if (fstat(STDIN_FILENO, &stdin_stat) == -1) {
		perror("fstat");
		close(sock);
		return EXIT_FAILURE;
	}

	if (!S_ISFIFO(stdin_stat.st_mode)) {
	        dprintf(STDERR_FILENO, "fatal: stdin must be a pipe\n");
	        close(sock);
	        return EXIT_FAILURE;
	}

	int pipe_size = fcntl(STDIN_FILENO, F_GETPIPE_SZ);
	if (pipe_size == -1) {
		perror("fcntl");
		close(sock);
		return EXIT_FAILURE;
	}

	ssize_t bytes_spliced;
	while ((bytes_spliced = splice(STDIN_FILENO, NULL, sock, NULL, pipe_size, SPLICE_F_MORE)) > 0) {
	}

	if (bytes_spliced == -1) {
		perror("splice");
		close(sock);
		return EXIT_FAILURE;
	}

	close(sock);
	return EXIT_SUCCESS;
}
