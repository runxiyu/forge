# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

.PHONY: clean version.go man

CFLAGS = -Wall -Wextra -Werror -pedantic -std=c99 -D_GNU_SOURCE
MAN_PAGES = forge.5 hookc.1

forge: version.go hookc/*.c hookc/hookc man # TODO
	go mod vendor
	go build .

man: $(MAN_PAGES:%=man/%.html)

man/%.html: man/%
	mandoc -Thtml -O style=static/mandoc.css $< > $@

hookc/hookc:

version.go:
	printf 'package main\n\nconst VERSION = "%s"\n' `git describe --tags --always --dirty` > $@

clean:
	$(RM) forge version.go vendor

