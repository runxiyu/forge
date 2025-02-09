# Lindenii Forge

## Organization

URLs are like `https://forge.example.org/project_name/module_type/module_name`.

The available `module_type`s are:
* `repos` for version control
* `tickets` for ticket trackers
* `mail` for mailing lists

## Version-controled repos

Currently we only support Git.

### Merge requests

Each version-controlled repo ("main repo") has an area for merge requests
("MR"s) which may be optionally enabled. A MR is a request to merge 
changes from a Git ref ("source ref") into a branch in the main repo
("destination branch").

When creating a MR from the API, the Web interface, email, or SSH, it shall be
possible to create a special source ref, hosted in the main repo as a branch
with a branch name that begins with `merge_requests/`. Unsolicited pushes to
`merge_requests/` will automatically open a MR, returning instructions to edit
the description and manage the MR further via the standard error channel.

MR branches shall be synced to automatically created MR-specific mailing lists.
These mailing lists should have archives accessible via read-only IMAP, JMAP,
or something else that achieves a similar result.

## Ticket tracking

Ticket tracking works like todo.sr.ht, though we also intend to support
IMAP/JMAP/etc.

## Mailing lists

Mailing lists are not designed to handle patchsets. Patchsets should be send to
the corresponding repo, where they are turned into Lindenii patchsets.
