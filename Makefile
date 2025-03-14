forge: templates.ha *.ha
	hare build -o $@ .

templates.ha: templates/*.htmpl
	htmplgen -o $@ $^
