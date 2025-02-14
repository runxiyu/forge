# Lindenii Forge

**Work in progress.**

Lindenii Forge aims to be an uncomplicated yet featured software forge,
primarily designed for self-hosting by small organizations and individuals.

## Setup

* Clone the source code and build a binary with `go build`.
* Generate an SSH key pair with `ssh-keygen`.
* Create a PostgreSQL database and run `schema.sql`.
* Set up reverse proxies, etc., if desired.
* Copy `forge.scfg` to `/etc/lindenii/forge.scfg` or another reasonable
  location and edit appropriately.

## Organization

Misc URLs (for hosting stylesheets, scipts, the login page, etc.) are like
`https://forge.example.org/:/functionality...`.

Contentful URLs are like
`https://forge.example.org/group_name/subgroup_name.../:/module_type/module_name`.

The available `module_type`s are:

* `repos` for version control
* `tickets` for ticket trackers
* `mail` for mailing lists

Group and subgroup names may not be `:`.

## Protocols and user interfaces

The following should be roughly equivalent in functionality:

* Web interface
* Custom TLS-based API
* HTTP GraphQL API

The following shall function where they make sense:

* User-friendly SSH interface
* Email interface

## Features

### Git repos

* Viewing commits, diffs, and patches
* Viewing trees and files
* Viewing arbitrary hashed objects and refs
* Viewing tags

### Merge requests

Each version-controlled repo ("main repo") has an area for merge requests
("MR"s) which may be optionally enabled. A MR is a request to merge 
changes from a Git ref ("source ref") into a branch in the main repo
("destination branch").

When creating a MR from the API, the web interface, email, or SSH, it shall be
possible to create a special source ref, hosted in the main repo as a branch
with a branch name that begins with `merge_requests/`. Unsolicited pushes to
`merge_requests/` will automatically open a MR, returning instructions to edit
the description and manage the MR further via the standard error channel.

MR branches shall be synced to automatically created MR-specific mailing lists.
These mailing lists should have archives accessible via read-only IMAP, JMAP,
or something else that achieves a similar result.

### Ticket tracking

Ticket tracking works like todo.sr.ht, though we also intend to support
IMAP/JMAP/etc to view their archives.

Simple tracking should work for now. Directed acrylic graph and other
dependency mechanisms may be considered in the far future.

Should be possible to associated with MRs.

### Mailing lists

Mailing lists are not designed to handle patchsets. Patchsets should be send to
the corresponding repo, where they are turned into Lindenii patchsets.

Mailing list messages are expected to be plain text. A subset of markdown shall
be considered. No full-HTML emails are expected for normal traffic.

### CI

We generally prefer to have linters, deployment pipelines and such run on each
local developer's machine, for example as a pre-commit hook. However there are
reasons why CI might be useful. We plan to integrate a git.sr.ht-style CI in
the future.

### Authentication and authorization

Anonymous SSH read access should be possible for public repos. All other Git
access should be done via SSH public keys. We use a baked-in SSH
implementation.

The native API may be authenticated in the transport layer (e.g. TLS client
certificates or UNIX domain socket authentication), via passwords, and via
challenge-response mechanisms including SSH keys. SASL may be considered.

The web interface will have a dedicated login screen. Connections are set as
keepalive, and sessions are tracked across a kept-alive connection; optionally,
a user may click "remember me with a cookie" in the login screen.

PGP patch validation may be considered.

## License

We currently use
[CC0 1.0 Universal](https://creativecommons.org/publicdomain/zero/1.0/legalcode.txt),
which is
[not](https://www.gnu.org/licenses/license-list.html#CC0)
[ideal](https://opensource.org/faq#cc-zero)
because it expressly disclaims granting patent licenses.

We have also considered the [GNU AGPL](www.gnu.org/licenses/agpl-3.0.en.html),
but it has various caveats that we don't fully understand.

Expect licensing to change in the future, although it will stay Libre beyond
reasonable doubt.

## Contact

We hang out in [`#chat`](https://webirc.runxiyu.org/kiwiirc/#chat)
and [`#lindenii`](https://webirc.runxiyu.org/kiwiirc/#lindenii)
on [irc.runxiyu.org](https://irc.runxiyu.org).
