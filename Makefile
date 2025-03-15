forge: templates.ha *.ha
	hare build $(HAREFLAGS) -o $@ .

templates.ha: templates/*.htmpl
	htmplgen -o $@ $^
