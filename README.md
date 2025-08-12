# Lindenii Forge

**Work in progress.**

Lindenii Forge aims to be an uncomplicated yet featured software forge,
primarily designed for self-hosting by small organizations and individuals.

* [Upstream source repository](https://forge.lindenii.runxiyu.org/forge/-/repos/server/)
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
* Converting mailed patches to branches

## Planned features

* Further integration with mailing list workflows
* Further federated authentication
* Ticket trackers, discussions, RFCs
  * Web interface
  * Email integration with IMAP archives
* SSH API
* Email access
* CI system similar to builds.sr.ht

## License

We are currently using the
[GNU Affero General Public License version 3](https://www.gnu.org/licenses/agpl-3.0.html).

The forge software serves its own source at `/-/source/`.

## Contribute

Please submit patches by pushing to `contrib/...` in the official repo.

Alternatively, send email to
[`forge/-/repos/server@forge.lindenii.runxiyu.org`](mailto:forge%2F-%2Frepos%2Fserver@forge.lindenii.runxiyu.org).
Note that emailing patches is still experimental.

## Mirrors

We have several repo mirrors:

* [Official repo on our own instance of Lindenii Forge](https://forge.lindenii.org/forge/-/repos/server/)
* [The Lindenii Project's backup cgit](https://git.lindenii.org/forge.git/)
* [SourceHut](https://git.sr.ht/~runxiyu/forge/)
* [GitHub](https://github.com/runxiyu/forge/)

## Architecture

We have a mostly monolithic server `forged` written in Go. PostgreSQL is used
to store everything other than Git repositories.

Git repositories currently must be accessible via the local filesystem from
the machine running `forged`, since `forged` currently uses `go-git`, `git2d`
via UNIX domain sockets, and `git-upload-pack`/`git-receive-pack` subprocesses.
In the future, `git2d` will be expanded to support all operations, removing
our dependence on `git-upload-pack`/`git-receive-pack` and `go-git`; `git2d`
will also be extended to support remote IPC via a custom RPC protocol,
likely based on SCTP (with TLS via RFC 3436).

## `git2d`

`git2d` is a Git server daemon written in C, which uses `libgit2` to handle Git
operations.

```c
int cmd_index(git_repository * repo, struct bare_writer *writer);
int cmd_treeraw(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_resolve_ref(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_list_branches(git_repository * repo, struct bare_writer *writer);
int cmd_format_patch(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_merge_base(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_log(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_tree_list_by_oid(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_write_tree(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_blob_write(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_commit_tree_oid(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_commit_create(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_update_ref(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_commit_info(git_repository * repo, struct bare_reader *reader, struct bare_writer *writer);
int cmd_init_repo(const char *path, struct bare_reader *reader, struct bare_writer *writer);
```

We are planning to rewrite `git2d` in Hare, using
[`hare-git`](https://forge.lindenii.org/hare/-/repos/hare-git/) when it's ready.
