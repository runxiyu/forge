// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>

use strings;
use net::uri;

// The result, if not erroring out, must be freed with strings::freeall.
fn segments_from_path(s: str) ([]str | nomem | uri::invalid) = {
	let sp: []str = strings::split(s, "/")?;
	for (let i = 1z; i < len(sp); i += 1) {
		match (uri::percent_decode(sp[i])) {
		case let s: str =>
			sp[i - 1] = s;
		case uri::invalid =>
			strings::freeall(sp[.. i - 1]);
			return uri::invalid;
		};
	};
	return sp[.. len(sp) - 1];
};
