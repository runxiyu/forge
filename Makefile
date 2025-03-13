forge: main.ha templates.ha
	hare build -o $@ .

templates.ha: templates/*.htmpl
	htmplgen -o $@ $^
