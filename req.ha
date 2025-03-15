// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

use htmpl;
use io;
use net::http;
use net::uri;
use strings;

export fn handlereq(conn: io::handle, request: *http::request) (void | io::error | nomem | uri::invalid) = {
	htmpl::write(conn, "HTTP/1.1 200 OK\r\n")?;
	htmpl::write(conn, "Content-Type: text/html\r\n\r\n")?;
	let segments = segments_from_path(request.target.raw_path)?;
	defer strings::freeall(segments);

	let trailing_slash: bool = false;

	if (segments[len(segments) - 1] == "") {
		trailing_slash = true;
		free(segments[len(segments) - 1]);
		segments = segments[.. len(segments) - 1];
	};
	
	tp_index(conn, segments)?;
};
