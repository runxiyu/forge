use strings;
use net::uri;

// The result must be freed with strings::freeall;
fn segments_from_path(s: str) ([]str | nomem | net::uri::invalid) = {
	let sp: []str = strings::split(s, "/")?;
	for (let i = 1z; i < len(sp); i += 1)
		sp[i - 1] = net::uri::percent_decode(sp[i])?;
	return sp[.. len(sp) - 1];
};
