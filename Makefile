.PHONY: clean

forge: $(filter-out forge,$(wildcard *)) version.go
	go build -o $@

version.go:
	printf 'package main\nconst VERSION="%s"\n' $(shell git describe --tags --long --always --dirty) > $@

clean:
	$(RM) forge version.go

