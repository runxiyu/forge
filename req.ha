// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

use fmt;
use htmpl;
use io;
use net::http;
use net::uri;
use strconv;
use strings;

fn handlereq(conn: io::handle, request: *http::request) (void | io::error | nomem | uri::invalid) = {
	let segments = segments_from_path(request.target.raw_path)?;
	defer strings::freeall(segments);

	let trailing_slash: bool = false;

	if (segments[len(segments) - 1] == "") {
		trailing_slash = true;
		free(segments[len(segments) - 1]);
		segments = segments[.. len(segments) - 1];
	};

	if (len(segments) == 0) {
		start_response(conn, 200, "text/html")?;
		return tp_index(conn, segments);
	};
	
	if (segments[0] == ":") {
		if (len(segments) == 1) {
			start_response(conn, 404, "text/plain")?;
			fmt::fprintln(conn, "Error: Blank system endpoint")?;
			return;
		};

		switch (segments[1]) {
		case "static" =>
			if (len(segments) == 2) {
				start_response(conn, 404, "text/plain")?;
				fmt::fprintln(conn, "Error: Blank static endpoint")?;
				return;
			};
			let fs_segments = segments[2 ..];
			for (let fs_segment .. fs_segments) {
				if (strings::contains(fs_segment, "/")) {
					start_response(conn, 400, "text/plain")?;
					fmt::fprintln(conn, "Error: Slash found in filesystem path")?;
					return;
				};
			};
			start_response(conn, 501, "text/plain")?;
			fmt::fprintln(conn, "Not implemented yet")?;
		case =>
			start_response(conn, 404, "text/plain")?;
			fmt::fprintln(conn, "Error: Unknown system endpoint")?;
		};
	};
};

fn start_response(conn: io::handle, status: uint, content_type: str) (void | io::error | nomem) = { // TODO: add len and other headers
	fmt::fprint(conn, "HTTP/1.1 ")?;
	fmt::fprint(conn, strconv::utos(status))?;
	fmt::fprint(conn, " ")?;
	fmt::fprint(conn, http::status_reason(status))?;
	fmt::fprint(conn, "\r\n")?;
	fmt::fprint(conn, "Content-Type: ")?;
	fmt::fprint(conn, content_type)?;
	fmt::fprint(conn, "\r\n")?;
	fmt::fprint(conn, "\r\n")?;
};
