#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <string.h>
#include <fcntl.h>

int main(void) {
	int sock;
	struct sockaddr_un addr;
	const char *socket_path = getenv("LINDENII_FORGE_HOOKS_SOCKET_PATH");

	if (socket_path == NULL) {
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

	ssize_t bytes_spliced;
	while ((bytes_spliced = splice(STDIN_FILENO, NULL, sock, NULL, 1, SPLICE_F_MORE)) > 0) {
	}

	if (bytes_spliced == -1) {
		perror("splice");
		close(sock);
		return EXIT_FAILURE;
	}

	close(sock);
	return EXIT_SUCCESS;
}
