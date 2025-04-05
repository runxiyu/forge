#include "x.h"

bare_error
conn_read(void *buffer, void *dst, uint64_t sz)
{
	conn_io_t *io = buffer;
	ssize_t rsz = read(io->fd, dst, sz);
	return (rsz == (ssize_t) sz) ? BARE_ERROR_NONE : BARE_ERROR_READ_FAILED;
}

bare_error
conn_write(void *buffer, const void *src, uint64_t sz)
{
	conn_io_t *io = buffer;
	const uint8_t *data = src;
	uint64_t total = 0;

	while (total < sz) {
		ssize_t written = write(io->fd, data + total, sz - total);
		if (written < 0) {
			if (errno == EINTR)
				continue;
			return BARE_ERROR_WRITE_FAILED;
		}
		if (written == 0)
			break;
		total += written;
	}

	return (total == sz) ? BARE_ERROR_NONE : BARE_ERROR_WRITE_FAILED;
}
