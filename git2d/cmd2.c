/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

int cmd_treeraw(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	/* Path */
	char path[4096] = { 0 };
	int err = bare_get_data(reader, (uint8_t *) path, sizeof(path) - 1);
	if (err != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	path[sizeof(path) - 1] = '\0';

	/* HEAD^{tree} */
	git_object *head_obj = NULL;
	err = git_revparse_single(&head_obj, repo, "HEAD^{tree}");
	if (err != 0) {
		bare_put_uint(writer, 4);
		return -1;
	}
	git_tree *tree = (git_tree *) head_obj;

	/* Path in tree */
	git_tree_entry *entry = NULL;
	git_otype objtype;
	if (strlen(path) == 0) {
		entry = NULL;
		objtype = GIT_OBJECT_TREE;
	} else {
		err = git_tree_entry_bypath(&entry, tree, path);
		if (err != 0) {
			bare_put_uint(writer, 3);
			git_tree_free(tree);
			return 0;
		}
		objtype = git_tree_entry_type(entry);
	}

	if (objtype == GIT_OBJECT_TREE) {
		/* Tree */
		git_object *tree_obj = NULL;
		if (entry == NULL) {
			tree_obj = (git_object *) tree;
		} else {
			err = git_tree_entry_to_object(&tree_obj, repo, entry);
			if (err != 0) {
				bare_put_uint(writer, 7);
				goto cleanup;
			}
		}
		git_tree *subtree = (git_tree *) tree_obj;

		size_t count = git_tree_entrycount(subtree);
		bare_put_uint(writer, 0);
		bare_put_uint(writer, 1);
		bare_put_uint(writer, count);
		for (size_t i = 0; i < count; i++) {
			const git_tree_entry *subentry = git_tree_entry_byindex(subtree, i);
			const char *name = git_tree_entry_name(subentry);
			git_otype type = git_tree_entry_type(subentry);
			uint32_t mode = git_tree_entry_filemode(subentry);

			uint8_t entry_type = 0;
			uint64_t size = 0;

			if (type == GIT_OBJECT_TREE) {
				entry_type = 1;
			} else if (type == GIT_OBJECT_BLOB) {
				entry_type = 2;

				git_object *subobj = NULL;
				if (git_tree_entry_to_object(&subobj, repo, subentry) == 0) {
					git_blob *b = (git_blob *) subobj;
					size = git_blob_rawsize(b);
					git_blob_free(b);
				}
			}

			bare_put_uint(writer, entry_type);
			bare_put_uint(writer, mode);
			bare_put_uint(writer, size);
			bare_put_data(writer, (const uint8_t *)name, strlen(name));
		}
		if (entry != NULL) {
			git_tree_free(subtree);
		}
	} else if (objtype == GIT_OBJECT_BLOB) {
		/* Blob */
		git_object *blob_obj = NULL;
		err = git_tree_entry_to_object(&blob_obj, repo, entry);
		if (err != 0) {
			bare_put_uint(writer, 7);
			goto cleanup;
		}
		git_blob *blob = (git_blob *) blob_obj;
		const void *content = git_blob_rawcontent(blob);
		if (content == NULL) {
			bare_put_uint(writer, 8);
			git_blob_free(blob);
			goto cleanup;
		}
		bare_put_uint(writer, 0);
		bare_put_uint(writer, 2);
		bare_put_data(writer, content, git_blob_rawsize(blob));
		git_blob_free(blob);
	} else {
		/* Unknown */
		bare_put_uint(writer, -1);
	}

 cleanup:
	if (entry != NULL)
		git_tree_entry_free(entry);
	git_tree_free(tree);
	return 0;
}
