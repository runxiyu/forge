.PHONY: clean version.go

CFLAGS = -Wall -Wextra -Werror -pedantic -std=c99 -D_GNU_SOURCE

forge: $(filter-out forge,$(wildcard *)) version.go git_hooks_client/*.c git_hooks_client/git_hooks_client
	go mod vendor
	go build -o $@

git_hooks_client/git_hooks_client:

version.go:
	printf 'package main\nconst VERSION="%s"\n' $(shell git describe --tags --always --dirty) > $@

clean:
	$(RM) forge version.go vendor

