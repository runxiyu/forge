# Lindenii Forge

[![builds.sr.ht status](https://builds.sr.ht/~runxiyu/forge.svg)](https://builds.sr.ht/~runxiyu/forge)

**Work in progress.**

Lindenii Forge aims to be an uncomplicated yet featured software forge,
primarily designed for self-hosting by small organizations and individuals.

* [Upstream source repository](https://forge.lindenii.runxiyu.org/lindenii/forge/:/repos/server/)
  ([backup](https://git.lindenii.runxiyu.org/forge.git/))
* [Website and documentation](https://lindenii.runxiyu.org/forge/)
* [Temporary issue tracker](https://todo.sr.ht/~runxiyu/forge)
* IRC [`#lindenii`](https://webirc.runxiyu.org/kiwiirc/#lindenii)
  on [irc.runxiyu.org](https://irc.runxiyu.org)\
  and [`#lindenii`](https://web.libera.chat/#lindenii)
  on [Libera.Chat](https://libera.chat)

## Implemented features

* Umambiguously parsable URL
* Groups and subgroups
* Repo hosting
* Push to `contrib/` branches to automatically create merge requests
* Basic federated authentication

## Planned features

* Integration with mailing list workflows
* Ticket trackers and discussions
  * Web interface
  * Email integration with IMAP archives
* SSH API
* Email access

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).

The forge software serves its own source at `/:/source/`.

## Contribute

Please submit patches by pushing to `contrib/...` in the official repo.

We have several repo mirrors:

* [Official repo on our own instance of Lindenii Forge](https://forge.lindenii.runxiyu.org/lindenii/forge/:/repos/server/)
* [The Lindenii Project's backup cgit](https://git.lindenii.runxiyu.org/forge.git/)
* [SourceHut](https://git.sr.ht/~runxiyu/forge/)
* [Codeberg](https://codeberg.org/lindenii/forge/)
* [GitHub](https://github.com/runxiyu/forge/)

## Hare implementation

There's a `hare` branch for an experimental implementation in the
[Hare](https://harelang.org) programming language. It's currently unused
because Hare isn't stable enough yet but we expect to pick it back up later.
