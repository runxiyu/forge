// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

use fmt;
use fs;
use htmpl;
use io;
use mime;
use net::http;
use net::uri;
use strconv;
use strings;

fn handlereq(conn: io::handle, request: *http::request) (void | io::error | nomem | fs::error) = {
	let segments = match(segments_from_path(request.target.raw_path)) {
	case let s: []str =>
		yield s;
	case uri::invalid =>
		start_response(conn, 400, "text/plain")?;
		fmt::fprintln(conn, "Invalid URI")?;
		return void;
	case nomem =>
		return nomem;
	case =>
		abort("unreachable");
	};
	defer strings::freeall(segments);

	let trailing_slash: bool = false;

	if (segments[len(segments) - 1] == "") {
		trailing_slash = true;
		free(segments[len(segments) - 1]);
		segments = segments[.. len(segments) - 1];
	};

	if (len(segments) == 0) {
		start_response(conn, 200, "text/html")?;
		return tp_index(conn);
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
			let fs_segment_path = strings::join("/", fs_segments...)?;
			defer free(fs_segment_path);

			let file = match (fs::open(static_fs as *fs::fs, fs_segment_path)) {
			case let f: io::handle => yield f;
			case fs::error =>
				start_response(conn, 500, "text/plain")?;
				fmt::fprintln(conn, "Filesystem error")?;
				return;
			};
			defer io::close(file)!;

			let ext = strings::rcut(fs_segments[len(fs_segments) - 1], ".").1;

			let mimetype = match (mime::lookup_ext(ext)) {
			case let m: *mime::mimetype => yield m.mime;
			case null => yield "application/octet-stream";
			};

			start_response(conn, 200, mimetype)?;
			io::copy(conn, file)?;

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
