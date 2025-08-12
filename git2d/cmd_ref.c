/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

static int write_oid(struct bare_writer *writer, const git_oid *oid)
{
	return bare_put_data(writer, oid->id, GIT_OID_RAWSZ) == BARE_ERROR_NONE ? 0 : -1;
}

int cmd_resolve_ref(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char type[32] = { 0 };
	char name[4096] = { 0 };
	if (bare_get_data(reader, (uint8_t *) type, sizeof(type) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	if (bare_get_data(reader, (uint8_t *) name, sizeof(name) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}

	git_oid oid = { 0 };
	int err = 0;

	if (type[0] == '\0') {
		git_object *obj = NULL;
		err = git_revparse_single(&obj, repo, "HEAD^{commit}");
		if (err != 0) {
			bare_put_uint(writer, 12);
			return -1;
		}
		git_commit *c = (git_commit *) obj;
		git_oid_cpy(&oid, git_commit_id(c));
		git_commit_free(c);
	} else if (strcmp(type, "commit") == 0) {
		err = git_oid_fromstr(&oid, name);
		if (err != 0) {
			bare_put_uint(writer, 12);
			return -1;
		}
	} else if (strcmp(type, "branch") == 0) {
		char fullref[4608];
		snprintf(fullref, sizeof(fullref), "refs/heads/%s", name);
		git_object *obj = NULL;
		err = git_revparse_single(&obj, repo, fullref);
		if (err != 0) {
			bare_put_uint(writer, 12);
			return -1;
		}
		git_commit *c = (git_commit *) obj;
		git_oid_cpy(&oid, git_commit_id(c));
		git_commit_free(c);
	} else if (strcmp(type, "tag") == 0) {
		char spec[4608];
		snprintf(spec, sizeof(spec), "refs/tags/%s^{commit}", name);
		git_object *obj = NULL;
		err = git_revparse_single(&obj, repo, spec);
		if (err != 0) {
			bare_put_uint(writer, 12);
			return -1;
		}
		git_commit *c = (git_commit *) obj;
		git_oid_cpy(&oid, git_commit_id(c));
		git_commit_free(c);
	} else {
		bare_put_uint(writer, 12);
		return -1;
	}

	bare_put_uint(writer, 0);
	return write_oid(writer, &oid);
}

int cmd_list_branches(git_repository *repo, struct bare_writer *writer)
{
	git_branch_iterator *it = NULL;
	int err = git_branch_iterator_new(&it, repo, GIT_BRANCH_LOCAL);
	if (err != 0) {
		bare_put_uint(writer, 13);
		return -1;
	}
	size_t count = 0;
	git_reference *ref;
	git_branch_t type;
	while (git_branch_next(&ref, &type, it) == 0) {
		count++;
		git_reference_free(ref);
	}
	git_branch_iterator_free(it);

	err = git_branch_iterator_new(&it, repo, GIT_BRANCH_LOCAL);
	if (err != 0) {
		bare_put_uint(writer, 13);
		return -1;
	}

	bare_put_uint(writer, 0);
	bare_put_uint(writer, count);
	while (git_branch_next(&ref, &type, it) == 0) {
		const char *name = NULL;
		git_branch_name(&name, ref);
		if (name == NULL)
			name = "";
		bare_put_data(writer, (const uint8_t *)name, strlen(name));
		git_reference_free(ref);
	}
	git_branch_iterator_free(it);
	return 0;
}
