# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

.PHONY: clean

CFLAGS = -Wall -Wextra -pedantic -std=c99 -D_GNU_SOURCE
MAN_PAGES = lindenii-forge.5 lindenii-forge-hookc.1 lindenii-forge.1 lindenii-forge-mail.5

VERSION = $(shell git describe --tags --always --dirty)
SOURCE_FILES = $(shell git ls-files)

forge: source.tar.gz hookc/hookc git2d/git2d $(MAN_PAGES:%=man/%.html) $(MAN_PAGES:%=man/%.txt) $(SOURCE_FILES)
	go build -o forge -ldflags="-X 'main.VERSION=$(VERSION)'" .

man/%.html: man/%
	mandoc -Thtml -O style=./mandoc.css $< > $@

man/%.txt: man/% utils/colb
	mandoc $< | ./utils/colb > $@

utils/colb:

hookc/hookc:

git2d/git2d: git2d/main.c git2d/bare.c git2d/utf8.c
	$(CC) $(CFLAGS) -o git2d/git2d $^ $(shell pkg-config --cflags --libs libgit2) -lpthread

clean:
	rm -rf forge vendor man/*.html man/*.txt utils/colb hookc/hookc git2d/git2d source.tar.gz */*.o

source.tar.gz: $(SOURCE_FILES)
	rm -f source.tar.gz
	go mod vendor
	git ls-files -z | xargs -0 tar -czf source.tar.gz vendor
