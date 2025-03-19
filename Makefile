# SPDX-License-Identifier: AGPL-3.0-only
# SPDX-FileContributor: Runxi Yu <https://runxiyu.org>

.PHONY: clean version.go

CFLAGS = -Wall -Wextra -Werror -pedantic -std=c99 -D_GNU_SOURCE

forge: $(filter-out forge,$(wildcard *)) version.go hookc/*.c hookc/hookc
	go mod vendor
	go build .

hookc/hookc:

version.go:
	printf 'package main\nconst VERSION="%s"\n' $(shell git describe --tags --always --dirty) > $@

clean:
	$(RM) forge version.go vendor

htmpl.go: htmpl/*
	gohtmplgen -o htmpl.go htmpl/*
