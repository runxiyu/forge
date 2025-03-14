# Lindenii Forge Development Notes

You will need the following dependencies:

- [hare](https://git.sr.ht/~sircmpwn/hare)
- [hare-http](https://git.sr.ht/~sircmpwn/hare-http) with
  [various patches](https://lists.sr.ht/~sircmpwn/hare-dev/patches?search=from%3Arunxiyu+prefix%3Ahare-http)
- [hare-htmpl](https://forge.runxiyu.org/hare/:/repos/hare-htmpl/)
  ([backup](https://git.sr.ht/~runxiyu/hare-htmpl))


Also, you'll need various horrible patches for `net::uri` before that gets fixed:

```
diff --git a/net/uri/+test.ha b/net/uri/+test.ha
index 345f41ee..63272d52 100644
--- a/net/uri/+test.ha
+++ b/net/uri/+test.ha
@@ -10,7 +10,7 @@ use net::ip;
 		uri {
 			scheme = "file",
 			host = "",
-			path = "/my/path/to/file",
+			raw_path = "/my/path/to/file",
 			...
 		},
 	)!;
@@ -19,7 +19,7 @@ use net::ip;
 		uri {
 			scheme = "http",
 			host = "harelang.org",
-			path = "/",
+			raw_path = "/",
 			...
 		},
 	)!;
@@ -38,7 +38,7 @@ use net::ip;
 			scheme = "ldap",
 			host = [13, 37, 73, 31]: ip::addr4,
 			port = 1234,
-			path = "/",
+			raw_path = "/",
 			...
 		},
 	)!;
@@ -47,7 +47,7 @@ use net::ip;
 		uri {
 			scheme = "http",
 			host = ip::parse("::1")!,
-			path = "/test",
+			raw_path = "/test",
 			...
 		},
 	)!;
@@ -58,7 +58,7 @@ use net::ip;
 		uri {
 			scheme = "urn",
 			host = "",
-			path = "example:animal:ferret:nose",
+			raw_path = "example:animal:ferret:nose",
 			...
 		},
 	)!;
@@ -67,7 +67,7 @@ use net::ip;
 		uri {
 			scheme = "mailto",
 			host = "",
-			path = "~sircmpwn/hare-dev@lists.sr.ht",
+			raw_path = "~sircmpwn/hare-dev@lists.sr.ht",
 			...
 		},
 	)!;
@@ -76,7 +76,7 @@ use net::ip;
 		uri {
 			scheme = "http",
 			host = "",
-			path = "/foo/bar",
+			raw_path = "/foo/bar",
 			...
 		},
 	)!;
@@ -85,7 +85,7 @@ use net::ip;
 		uri {
 			scheme = "http",
 			host = "",
-			path = "/",
+			raw_path = "/",
 			...
 		},
 	)!;
@@ -94,7 +94,7 @@ use net::ip;
 		uri {
 			scheme = "https",
 			host = "sr.ht",
-			path = "/projects",
+			raw_path = "/projects",
 			query = "search=%23risc-v&sort=longest-active",
 			fragment = "foo",
 			...
@@ -105,7 +105,7 @@ use net::ip;
 		uri {
 			scheme = "https",
 			host = "en.wiktionary.org",
-			path = "/wiki/おはよう",
+			raw_path = "/wiki/%E3%81%8A%E3%81%AF%E3%82%88%E3%81%86",
 			fragment = "Japanese",
 			...
 		}
@@ -135,11 +135,11 @@ use net::ip;
 
 @test fn percent_encoding() void = {
 	test_uri(
-		"https://git%2esr.ht/~sircmpw%6e/hare#Build%20status",
+		"https://git.sr.ht/~sircmpwn/hare#Build%20status",
 		uri {
 			scheme = "https",
 			host = "git.sr.ht",
-			path = "/~sircmpwn/hare",
+			raw_path = "/~sircmpwn/hare",
 			fragment = "Build status",
 			...
 		},
@@ -152,7 +152,7 @@ use net::ip;
 		uri {
 			scheme = "ldap",
 			host = ip::parse("2001:db8::7")!,
-			path = "/c=GB",
+			raw_path = "/c=GB",
 			query = "objectClass?one",
 			...
 		},
@@ -161,11 +161,11 @@ use net::ip;
 
 	// https://bugs.chromium.org/p/chromium/issues/detail?id=841105
 	test_uri(
-		"https://web-safety.net/..;@www.google.com:%3443",
+		"https://web-safety.net/..;@www.google.com:443",
 		uri {
 			scheme = "https",
 			host = "web-safety.net",
-			path = "/..;@www.google.com:443",
+			raw_path = "/..;@www.google.com:443",
 			...
 		},
 		"https://web-safety.net/..;@www.google.com:443",
@@ -180,6 +180,7 @@ fn test_uri(in: str, expected_uri: uri, expected_str: str) (void | invalid) = {
 	const u = parse(in)?;
 	defer finish(&u);
 
+
 	assert_str(u.scheme, expected_uri.scheme);
 	match (u.host) {
 	case let s: str =>
@@ -189,7 +190,7 @@ fn test_uri(in: str, expected_uri: uri, expected_str: str) (void | invalid) = {
 	};
 	assert(u.port == expected_uri.port);
 	assert_str(u.userinfo, expected_uri.userinfo);
-	assert_str(u.path, expected_uri.path);
+	assert_str(u.raw_path, expected_uri.raw_path);
 	assert_str(u.query, expected_uri.query);
 	assert_str(u.fragment, expected_uri.fragment);
 
diff --git a/net/uri/fmt.ha b/net/uri/fmt.ha
index 48a43f24..07cb3f7b 100644
--- a/net/uri/fmt.ha
+++ b/net/uri/fmt.ha
@@ -20,9 +20,9 @@ use strings;
 // query      = *( pchar / "/" / "?" )
 // fragment   = *( pchar / "/" / "?" )
 
-def unres_host: str = "-._~!$&'()*+,;=";
-def unres_query_frag: str = "-._~!$&'()*+,;=:@/?";
-def unres_path: str = "-._~!$&'()*+,;=:@/";
+export def unres_host: str = "-._~!$&'()*+,;=";
+export def unres_query_frag: str = "-._~!$&'()*+,;=:@/?";
+export def unres_path: str = "-._~!$&'()*+,;=:@/";
 
 // Writes a formatted [[uri]] to an [[io::handle]]. Returns the number of bytes
 // written.
@@ -63,10 +63,10 @@ export fn fmt(out: io::handle, u: *const uri) (size | io::error) = {
 	if (u.port != 0) {
 		n += fmt::fprintf(out, ":{}", u.port)?;
 	};
-	if (has_host && len(u.path) > 0 && !strings::hasprefix(u.path, '/')) {
+	if (has_host && len(u.raw_path) > 0 && !strings::hasprefix(u.raw_path, '/')) {
 		n += fmt::fprint(out, "/")?;
 	};
-	n += percent_encode(out, u.path, unres_path)?;
+	n += memio::concat(out, u.raw_path)?;
 	if (len(u.query) > 0) {
 		// Always percent-encoded, see parse and encodequery/decodequery
 		n += fmt::fprintf(out, "?{}", u.query)?;
@@ -92,7 +92,7 @@ fn fmtaddr(out: io::handle, addr: ip::addr) (size | io::error) = {
 	return n;
 };
 
-fn percent_encode(out: io::handle, src: str, allowed: str) (size | io::error) = {
+export fn percent_encode(out: io::handle, src: str, allowed: str) (size | io::error) = {
 	let iter = strings::iter(src);
 	let n = 0z;
 	for (let r => strings::next(&iter)) {
diff --git a/net/uri/parse.ha b/net/uri/parse.ha
index f2522c01..e108bd75 100644
--- a/net/uri/parse.ha
+++ b/net/uri/parse.ha
@@ -22,10 +22,10 @@ export fn parse(in: str) (uri | invalid) = {
 	defer if (!success) free(scheme);
 
 	// Determine hier-part variant
-	let path = "";
+	let raw_path = "";
 	let authority: ((str | ip::addr6), u16, str) = ("", 0u16, "");
 	defer if (!success) {
-		free(path);
+		free(raw_path);
 		free_host(authority.0);
 		free(authority.2);
 	};
@@ -50,7 +50,7 @@ export fn parse(in: str) (uri | invalid) = {
 						case '/' =>
 							// path-absolute
 							strings::prev(&in);
-							path = parse_path(&in,
+							raw_path = parse_path(&in,
 								path_mode::ABSOLUTE)?;
 						case =>
 							return invalid;
@@ -61,17 +61,17 @@ export fn parse(in: str) (uri | invalid) = {
 					// path-absolute
 					strings::prev(&in); // return current token
 					strings::prev(&in); // return leading slash
-					path = parse_path(&in, path_mode::ABSOLUTE)?;
+					raw_path = parse_path(&in, path_mode::ABSOLUTE)?;
 				};
 			case =>
 				// path-absolute (just '/')
 				strings::prev(&in); // return leading slash
-				path = parse_path(&in, path_mode::ABSOLUTE)?;
+				raw_path = parse_path(&in, path_mode::ABSOLUTE)?;
 			};
 		case =>
 			// path-rootless
 			strings::prev(&in);
-			path = parse_path(&in, path_mode::ROOTLESS)?;
+			raw_path = parse_path(&in, path_mode::ROOTLESS)?;
 		};
 	case => void; // path-empty
 	};
@@ -118,7 +118,7 @@ export fn parse(in: str) (uri | invalid) = {
 		port = authority.1,
 		userinfo = authority.2,
 
-		path = path,
+		raw_path = raw_path,
 		query = query,
 		fragment = fragment,
 	};
@@ -274,7 +274,7 @@ fn parse_path(in: *strings::iterator, mode: path_mode) (str | invalid) = {
 		};
 	};
 
-	return percent_decode(strings::slice(&copy, in));
+	return strings::dup(strings::slice(&copy, in))!;
 };
 
 fn parse_query(in: *strings::iterator) (str | invalid) = {
@@ -323,13 +323,14 @@ fn parse_port(in: *strings::iterator) (u16 | invalid) = {
 	};
 };
 
-fn percent_decode(s: str) (str | invalid) = {
+// must be freed by caller
+export fn percent_decode(s: str) (str | invalid) = {
 	let buf = memio::dynamic();
 	percent_decode_static(&buf, s)?;
 	return memio::string(&buf)!;
 };
 
-fn percent_decode_static(out: io::handle, s: str) (void | invalid) = {
+export fn percent_decode_static(out: io::handle, s: str) (void | invalid) = {
 	let iter = strings::iter(s);
 	let tmp = memio::dynamic();
 	defer io::close(&tmp)!;
diff --git a/net/uri/uri.ha b/net/uri/uri.ha
index 623ffafb..3b7b7c4c 100644
--- a/net/uri/uri.ha
+++ b/net/uri/uri.ha
@@ -12,7 +12,7 @@ export type uri = struct {
 	port: u16,
 	userinfo: str,
 
-	path: str,
+	raw_path: str,
 	query: str,
 	fragment: str,
 };
@@ -31,7 +31,7 @@ export fn dup(u: *uri) uri = {
 		port = u.port,
 		userinfo = strings::dup(u.userinfo)!,
 
-		path = strings::dup(u.path)!,
+		raw_path = strings::dup(u.raw_path)!,
 		query = strings::dup(u.query)!,
 		fragment = strings::dup(u.fragment)!,
 	};
@@ -46,7 +46,7 @@ export fn finish(u: *uri) void = {
 	case => void;
 	};
 	free(u.userinfo);
-	free(u.path);
+	free(u.raw_path);
 	free(u.query);
 	free(u.fragment);
 };
```
