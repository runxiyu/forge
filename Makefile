.PHONY: clean version.go

forge: $(filter-out forge,$(wildcard *)) version.go
	go mod vendor
	go build -o $@
	$(RM) -r vendor

version.go:
	printf 'package main\nconst VERSION="%s"\n' $(shell git describe --tags --long --always --dirty) > $@

clean:
	$(RM) forge version.go

