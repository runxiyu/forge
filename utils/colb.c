/*-
 * SPDX-License-Identifier: GPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2021 June McEnroe <june@causal.agency>
 */

#include <locale.h>
#include <stdio.h>
#include <wchar.h>

int
main(void)
{
	wint_t		next, prev = WEOF;

	setlocale(LC_CTYPE, "C.UTF-8");

	while (WEOF != (next = getwchar())) {
		if (next == L'\b') {
			prev = WEOF;
		} else {
			if (prev != WEOF)
				putwchar(prev);
			prev = next;
		}
	}
	if (prev != WEOF)
		putwchar(prev);
}
