// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
// Adapted from template by Willow Barraco <contact@willowbarraco.fr>

use getopt;
use log;
use net;
use net::ip;
use net::http;
use net::dial;
use os;
use memio;
use io;
use fmt;
use bufio;
use strings;

const usage: [_]getopt::help = [
	"HTTP server",
	('a', "address", "listened address (ex: 127.0.0.1:8080)")
];

export fn main() void = {
	const cmd = getopt::parse(os::args, usage...);
	defer getopt::finish(&cmd);

	let port: u16 = 8080;
	let ip_addr: ip::addr4 = [127, 0, 0, 1];

	for (let opt .. cmd.opts) {
		switch (opt.0) {
		case 'a' =>
			match (dial::splitaddr(opt.1, "")) {
			case let value: (str, u16) =>
				ip_addr = ip::parsev4(value.0)!;
				port = value.1;
			case dial::invalid_address =>
				abort("invalid address");
			};
		case => abort(); // unreachable
		};
	};

	const server = match (http::listen(ip_addr, port)) {
	case let this: *http::server =>
		yield this;
	case net::error => abort("failure while listening");
	};
	defer http::server_finish(server);

	for (true) {
		const serv_req = match (http::serve(server)) {
		case let this: *http::server_request =>
			yield this;
		case net::error => abort("failure while serving");
		};
		defer http::serve_finish(serv_req);

		match (handlereq(serv_req.socket, &serv_req.request)) {
		case void => yield;
		case io::error => log::println("error while handling request");
		};
	};
};

export fn handlereq(conn: io::handle, request: *http::request) (void | io::error)  = {
	fmt::fprint(conn, "HTTP/1.1 200 OK\r\n")?;
	fmt::fprint(conn, "Content-Type: text/plain\r\n\r\n")?;

	fmt::fprintln(conn, "Hi there!")?;
};
