#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <string.h>

int main() {
	int sock;
	struct sockaddr_un addr;
	const char *message = "hi";
	const char *socket_path = getenv("LINDENII_FORGE_HOOKS_SOCKET_PATH");

	if (socket_path == NULL) {
		exit(EXIT_FAILURE);
	}

	sock = socket(AF_UNIX, SOCK_STREAM, 0);
	if (sock == -1) {
		perror("socket");
		exit(EXIT_FAILURE);
	}

	memset(&addr, 0, sizeof(struct sockaddr_un));
	addr.sun_family = AF_UNIX;
	strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path) - 1);

	if (connect(sock, (struct sockaddr *)&addr, sizeof(struct sockaddr_un)) == -1) {
		perror("connect");
		close(sock);
		exit(EXIT_FAILURE);
	}

	if (send(sock, message, strlen(message), 0) == -1) {
		perror("send");
		close(sock);
		exit(EXIT_FAILURE);
	}

	close(sock);

	return 0;
}
