# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
#
# TODO: This Makefile utilizes a lot of GNU extensions. Some of them are
# unfortunately difficult to avoid as POSIX Make's pattern rules are not
# sufficiently expressive. This needs to be fixed sometime (or we might move to
# some other build system).
#

.PHONY: clean

CFLAGS = -Wall -Wextra -pedantic -std=c99 -D_GNU_SOURCE
MAN_PAGES = lindenii-forge.5 lindenii-forge-hookc.1 lindenii-forge.1 lindenii-forge-mail.5

VERSION = $(shell git describe --tags --always --dirty)
SOURCE_FILES = $(shell git ls-files)

forge: source.tar.gz hookc/hookc git2d/git2d $(SOURCE_FILES)
	CGO_ENABLED=0 go build -o forge -ldflags '-extldflags "-f no-PIC -static" -X "go.lindenii.runxiyu.org/forge.version=$(VERSION)"' -tags 'osusergo netgo static_build' ./cmd/forge

utils/colb:

hookc/hookc:

git2d/git2d: git2d/*.c
	$(CC) $(CFLAGS) -o git2d/git2d $^ $(shell pkg-config --cflags --libs libgit2) -lpthread

clean:
	rm -rf forge utils/colb hookc/hookc git2d/git2d source.tar.gz */*.o

source.tar.gz: $(SOURCE_FILES)
	rm -f source.tar.gz
	git ls-files -z | xargs -0 tar -czf source.tar.gz
