// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
// Adapted from template by Willow Barraco <contact@willowbarraco.fr>

use fs;
use getopt;
use log;
use net;
use net::dial;
use net::http;
use net::ip;
use net::tcp;
use net::uri;
use os;
use memio;
use io;
use fmt;
use bufio;

const usage: [_]getopt::help = [
	"Lindenii Forge Server",
	('c', "config", "path to configuration file")
];

let static_fs: nullable *fs::fs = null;

export fn main() void = {
	const cmd = getopt::parse(os::args, usage...);
	defer getopt::finish(&cmd);

	let port: u16 = 8080;
	let ip_addr: ip::addr4 = [127, 0, 0, 1];

	for (let opt .. cmd.opts) {
		switch (opt.0) {
		case 'c' => yield; // TODO: actually handle the config
		case => abort("unreachable");
		};
	};

	static_fs = os::diropen("static")!;

	const server = match (http::listen(ip_addr, port, net::tcp::reuseport, net::tcp::reuseaddr)) {
	case let this: *http::server =>
		yield this;
	case => abort("failure while listening");
	};
	defer http::server_finish(server);

	for (true) {
		const serv_req = match (http::serve(server)) {
		case let this: *http::server_request =>
			yield this;
		case =>
			log::println("failure while serving");
			continue;
		};
		defer http::serve_finish(serv_req);

		match (handlereq(serv_req.socket, &serv_req.request)) {
		case void => yield;
		case => log::println("error while handling request");
		};
	};
};
