forge: .version.ha .templates.ha *.ha
	hare build $(HAREFLAGS) -o $@ .

.templates.ha: templates/*.htmpl
	htmplgen -o $@ $^

.version.ha:
	printf 'def VERSION="%s";\n' $(shell git describe --tags --always --dirty) > $@

.PHONY: version.ha
