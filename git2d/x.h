/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include <err.h>
#include <errno.h>
#include <git2.h>
#include <pthread.h>
#include <signal.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/un.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "bare.h"

#ifndef X_H
#define X_H

typedef struct {
	int fd;
} conn_io_t;


bare_error conn_read(void *buffer, void *dst, uint64_t sz);
bare_error conn_write(void *buffer, const void *src, uint64_t sz);

void * session(void *_conn);

int cmd1(git_repository *repo, struct bare_writer *writer);

#endif // X_H
