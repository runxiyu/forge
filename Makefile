# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

.PHONY: clean version.go man source.tar.gz

CFLAGS = -Wall -Wextra -Werror -pedantic -std=c99 -D_GNU_SOURCE
MAN_PAGES = lindenii-forge.5 lindenii-forge-hookc.1 lindenii-forge.1 lindenii-forge-mail.5

forge: source.tar.gz version.go hookc/*.c hookc/hookc man # TODO
	go build .

man: $(MAN_PAGES:%=man/%.html) $(MAN_PAGES:%=man/%.txt)

man/%.html: man/%
	mandoc -Thtml -O style=./mandoc.css $< > $@

man/%.txt: man/% utils/colb
	mandoc $< | ./utils/colb > $@

utils/colb:

hookc/hookc:

version.go:
	printf 'package main\n\nconst VERSION = "%s"\n' `git describe --tags --always --dirty` > $@

clean:
	$(RM) forge version.go vendor

source.tar.gz:
	rm -f source.tar.gz
	go mod vendor
	git ls-files -z | xargs -0 tar -czf source.tar.gz vendor
