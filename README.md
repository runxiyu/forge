# Lindenii Forge

**Work in progress.**

Lindenii Forge aims to be an uncomplicated yet featured software forge,
primarily designed for self-hosting by small organizations and individuals.

The Lindenii project itself
[runs an instance](https://forge.lindenii.runxiyu.org/),
where the
[official source repository of Lindenii Forge](https://forge.lindenii.runxiyu.org/lindenii/:/repos/forge/)
is located.

## Setup

* Clone the source code and build a binary with `make`.
  (You will need a Go toolchain, a C compiler, and GNU Make.)
* Generate an SSH key pair with `ssh-keygen`.
* Create a PostgreSQL database and run `schema.sql`.
* Set up reverse proxies, etc., if desired.
* Copy `forge.scfg` to `/etc/lindenii/forge.scfg` or another reasonable
  location and edit appropriately.

When email integration is ready and you wish to use email integration, you will
need to set up
[Lindenii Mail Daemon](https://forge.lindenii.runxiyu.org/lindenii/:/repos/maild/)
and configure it accordingly.

## Organization

Misc URLs (for hosting stylesheets, scripts, the login page, etc.) are like
`https://forge.example.org/:/functionality...`.

Contentful URLs are like
`https://forge.example.org/group_name/subgroup_name.../:/module_type/module_name`.

The available `module_type`s are:

* `repos` for version control
* `tickets` for ticket trackers
* `mail` for mailing lists

Group and subgroup names may not be `:`.

While this syntax looks a bit odd, it makes it possible to unambiguously
identify where subgroups end and modules begin, without needing to touch the
database.

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

* Viewing commits and patches
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
with a branch name that begins with `contrib/`. Unsolicited pushes to
`contrib/` automatically open a MR, returning instructions to edit the
description and manage the MR further via the standard error channel.

MR branches shall be synced to automatically created MR-specific mailing lists.
These mailing lists should have archives accessible via read-only IMAP, JMAP,
or something else that achieves a similar result. Merge requests are presented
as what would be produced from git-send-email. It should also be possible for
people to perform code reviews via email by interwoven a quoted patch with
replies, and all these replies should be synced to the main MR database.

This is probably the most unique feature of Lindenii Forge: while the general
structure is similar to Forgejo-style pull requests, MRs are automatically
created based on branch namespaces (similar to Gerrit in some respects) and are
synced to mailing lists. This allows users to submit MRs via email, or git
push, or the Web interface, or the API, to support developers with different
workflows.

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
reasons why CI might be useful. We plan to integrate a builds.sr.ht-style CI in
the future.

### Authentication and authorization

Anonymous SSH and HTTPS read access is possible for public repos. Git write
access is done via SSH public keys. We use a baked-in SSH implementation.

The native API may be authenticated in the transport layer (e.g. TLS client
certificates or UNIX domain socket authentication), via passwords, and via
challenge-response mechanisms including SSH keys. SASL may be considered.

The web interface has a dedicated login screen. Logins are remembered with
cookies. A dropdown box shall be available to select the time the user wishes
to remain logged in.

Commit validation based on SSH signature validation will be implemented.

### Federated authentication

We probably won't fully support ForgeFed because it's way too bloated. However,
some type of federated authentication may be considered.

Since Forgejo, SourceHut, and GitHub all publicly serve their users' SSH keys,
people who submit merge requests by pushing via SSH into the branch namespace
may link their SSH keys to an identity on an external forge. We will also serve
users' SSH keys, but it would be opt-in.

OpenID Connect will likely not be supported.

This plan is subject to change.

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).

The forge software serves its own source at `/:/source/`.

## Support and development

* We hang out in [`#chat`](https://webirc.runxiyu.org/kiwiirc/#chat)
  and [`#lindenii`](https://webirc.runxiyu.org/kiwiirc/#lindenii)
  on [irc.runxiyu.org](https://irc.runxiyu.org).
  The latter is bridged to [`#lindenii`](https://web.libera.chat/#lindenii)
  on [Libera.Chat](https://libera.chat).
* Issues and pull requests may be submitted on the Codeberg mirror before our
  own ticket tracking and merge request systems are ready. However, anonymous
  pushes to `contrib/` branches should already be possible.
* We have several Git repo mirrors on a few places:
  * [Lindenii Forge itself](https://forge.lindenii.runxiyu.org/lindenii/:/repos/forge/)
  * [The Lindenii Project's cgit](https://git.lindenii.runxiyu.org/forge.git/)
  * [SourceHut](https://git.sr.ht/~runxiyu/forge/)
  * [Codeberg](https://codeberg.org/lindenii/forge/)
  * [GitHub](https://github.com/runxiyu/forge/)

## Code style

We follow the Lindenii Project's general code style, which has a few important
deviations from what most people may be used to:

* We used tabs for indentation everywhere (Go, HTML, CSS, JS, SQL, etc.)
* All names are in `snake_case`; exported identifiers in Go use
  `Capitalized_snake_case`. Type identifiers end with `_t`.
