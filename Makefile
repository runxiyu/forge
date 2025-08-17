# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
#
# TODO: This Makefile utilizes a lot of GNU extensions. Some of them are
# unfortunately difficult to avoid as POSIX Make's pattern rules are not
# sufficiently expressive. This needs to be fixed sometime (or we might move to
# some other build system).
#

.PHONY: clean all

CFLAGS = -Wall -Wextra -pedantic -std=c99 -D_GNU_SOURCE

all: dist/forged dist/git2d dist/hookc

dist/forged: $(shell git ls-files forged)
	mkdir -p dist
	sqlc -f forged/sqlc.yaml generate
	CGO_ENABLED=0 go build -o dist/forged -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' ./forged

dist/git2d: $(wildcard git2d/*.c)
	mkdir -p dist
	$(CC) $(CFLAGS) -o dist/git2d $^ $(shell pkg-config --cflags --libs libgit2) -lpthread

dist/hookc: $(wildcard hookc/*.c)
	mkdir -p dist
	$(CC) $(CFLAGS) -o dist/hookc $^

clean:
	rm -rf dist

