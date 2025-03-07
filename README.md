# Lindenii Forge

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

## Planned features

* Umambiguously parsable URL
* Groups and subgroups
* Repo hosting
* Merge requests
  * Push to `contrib/` branches to automatically create MRs
  * Integration with traditional mailing list workflows
* Ticket trackers
  * Email integration with IMAP archives
  * Web interface
* Discussions
  * Email integration with IMAP archives
  * Web interface
* Multiple user interfaces: web, SSH, custom API
* Federated authentication

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).

The forge software serves its own source at `/:/source/`.

## Support and development

Please submit patches by pushing to `contrib/...` in the official repo.

We have several Git repo mirrors on a few places:
* [Lindenii Forge itself](https://forge.lindenii.runxiyu.org/lindenii/forge/:/repos/server/)
* [The Lindenii Project's cgit](https://git.lindenii.runxiyu.org/forge.git/)
* [SourceHut](https://git.sr.ht/~runxiyu/forge/)
* [Codeberg](https://codeberg.org/lindenii/forge/)
* [GitHub](https://github.com/runxiyu/forge/)

## Code style

We follow the Lindenii Project's general code style, which has a few important
deviations from what most people may be used to:

* We used tabs for indentation everywhere (Go, C, HTML, CSS, JS, SQL, etc.)
* All names are in `snake_case`; exported identifiers in Go use
  `Capitalized_snake_case`. Type identifiers end with `_t`.
* We avoid `:=` in Go, and prefer to use `var`.
