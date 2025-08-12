/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

int cmd_tree_list_by_oid(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char hex[64] = { 0 };
	if (bare_get_data(reader, (uint8_t *) hex, sizeof(hex) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid oid;
	if (git_oid_fromstr(&oid, hex) != 0) {
		bare_put_uint(writer, 4);
		return -1;
	}
	git_tree *tree = NULL;
	if (git_tree_lookup(&tree, repo, &oid) != 0) {
		bare_put_uint(writer, 4);
		return -1;
	}
	size_t count = git_tree_entrycount(tree);
	bare_put_uint(writer, 0);
	bare_put_uint(writer, count);
	for (size_t i = 0; i < count; i++) {
		const git_tree_entry *e = git_tree_entry_byindex(tree, i);
		const char *name = git_tree_entry_name(e);
		uint32_t mode = git_tree_entry_filemode(e);
		const git_oid *id = git_tree_entry_id(e);
		bare_put_uint(writer, mode);
		bare_put_data(writer, (const uint8_t *)name, strlen(name));
		bare_put_data(writer, id->id, GIT_OID_RAWSZ);
	}
	git_tree_free(tree);
	return 0;
}

int cmd_write_tree(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	uint64_t count = 0;
	if (bare_get_uint(reader, &count) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_treebuilder *bld = NULL;
	if (git_treebuilder_new(&bld, repo, NULL) != 0) {
		bare_put_uint(writer, 15);
		return -1;
	}
	for (uint64_t i = 0; i < count; i++) {
		uint64_t mode = 0;
		if (bare_get_uint(reader, &mode) != BARE_ERROR_NONE) {
			git_treebuilder_free(bld);
			bare_put_uint(writer, 11);
			return -1;
		}
		char name[4096] = { 0 };
		if (bare_get_data(reader, (uint8_t *) name, sizeof(name) - 1) != BARE_ERROR_NONE) {
			git_treebuilder_free(bld);
			bare_put_uint(writer, 11);
			return -1;
		}
		uint8_t idraw[GIT_OID_RAWSZ] = { 0 };
		if (bare_get_fixed_data(reader, idraw, GIT_OID_RAWSZ) != BARE_ERROR_NONE) {
			git_treebuilder_free(bld);
			bare_put_uint(writer, 11);
			return -1;
		}
		git_oid id;
		memcpy(id.id, idraw, GIT_OID_RAWSZ);
		git_filemode_t fm = (git_filemode_t) mode;
		if (git_treebuilder_insert(NULL, bld, name, &id, fm) != 0) {
			git_treebuilder_free(bld);
			bare_put_uint(writer, 15);
			return -1;
		}
	}
	git_oid out;
	if (git_treebuilder_write(&out, bld) != 0) {
		git_treebuilder_free(bld);
		bare_put_uint(writer, 15);
		return -1;
	}
	git_treebuilder_free(bld);
	bare_put_uint(writer, 0);
	bare_put_data(writer, out.id, GIT_OID_RAWSZ);
	return 0;
}

int cmd_blob_write(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	uint64_t sz = 0;
	if (bare_get_uint(reader, &sz) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	uint8_t *data = (uint8_t *) malloc(sz);
	if (!data) {
		bare_put_uint(writer, 15);
		return -1;
	}
	if (bare_get_fixed_data(reader, data, sz) != BARE_ERROR_NONE) {
		free(data);
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid oid;
	if (git_blob_create_frombuffer(&oid, repo, data, sz) != 0) {
		free(data);
		bare_put_uint(writer, 15);
		return -1;
	}
	free(data);
	bare_put_uint(writer, 0);
	bare_put_data(writer, oid.id, GIT_OID_RAWSZ);
	return 0;
}
