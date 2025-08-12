/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

static int append_buf(char **data, size_t *len, size_t *cap, const char *src, size_t n)
{
	if (n == 0)
		return 0;
	size_t need = *len + n;
	if (need > *cap) {
		size_t newcap = *cap ? *cap * 2 : 256;
		while (newcap < need)
			newcap *= 2;
		char *p = (char *)realloc(*data, newcap);
		if (!p)
			return -1;
		*data = p;
		*cap = newcap;
	}
	memcpy(*data + *len, src, n);
	*len += n;
	return 0;
}

int cmd_commit_tree_oid(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char hex[64] = { 0 };
	if (bare_get_data(reader, (uint8_t *) hex, sizeof(hex) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid oid;
	if (git_oid_fromstr(&oid, hex) != 0) {
		bare_put_uint(writer, 14);
		return -1;
	}
	git_commit *commit = NULL;
	if (git_commit_lookup(&commit, repo, &oid) != 0) {
		bare_put_uint(writer, 14);
		return -1;
	}
	git_tree *tree = NULL;
	if (git_commit_tree(&tree, commit) != 0) {
		git_commit_free(commit);
		bare_put_uint(writer, 14);
		return -1;
	}
	const git_oid *toid = git_tree_id(tree);
	bare_put_uint(writer, 0);
	bare_put_data(writer, toid->id, GIT_OID_RAWSZ);
	git_tree_free(tree);
	git_commit_free(commit);
	return 0;
}

int cmd_commit_create(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char treehex[64] = { 0 };
	if (bare_get_data(reader, (uint8_t *) treehex, sizeof(treehex) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid tree_oid;
	if (git_oid_fromstr(&tree_oid, treehex) != 0) {
		bare_put_uint(writer, 15);
		return -1;
	}
	uint64_t pcnt = 0;
	if (bare_get_uint(reader, &pcnt) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_commit **parents = NULL;
	if (pcnt > 0) {
		parents = (git_commit **) calloc(pcnt, sizeof(git_commit *));
		if (!parents) {
			bare_put_uint(writer, 15);
			return -1;
		}
		for (uint64_t i = 0; i < pcnt; i++) {
			char phex[64] = { 0 };
			if (bare_get_data(reader, (uint8_t *) phex, sizeof(phex) - 1) != BARE_ERROR_NONE) {
				bare_put_uint(writer, 11);
				goto fail;
			}
			git_oid poid;
			if (git_oid_fromstr(&poid, phex) != 0) {
				bare_put_uint(writer, 15);
				goto fail;
			}
			if (git_commit_lookup(&parents[i], repo, &poid) != 0) {
				bare_put_uint(writer, 15);
				goto fail;
			}
		}
	}
	char aname[512] = { 0 };
	char aemail[512] = { 0 };
	if (bare_get_data(reader, (uint8_t *) aname, sizeof(aname) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		goto fail;
	}
	if (bare_get_data(reader, (uint8_t *) aemail, sizeof(aemail) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		goto fail;
	}
	int64_t when = 0;
	int64_t tzoff = 0;
	if (bare_get_i64(reader, &when) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		goto fail;
	}
	if (bare_get_i64(reader, &tzoff) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		goto fail;
	}
	char *message = NULL;
	{
		uint64_t msz = 0;
		if (bare_get_uint(reader, &msz) != BARE_ERROR_NONE) {
			bare_put_uint(writer, 11);
			goto fail;
		}
		message = (char *)malloc(msz + 1);
		if (!message) {
			bare_put_uint(writer, 15);
			goto fail;
		}
		if (bare_get_fixed_data(reader, (uint8_t *) message, msz) != BARE_ERROR_NONE) {
			free(message);
			bare_put_uint(writer, 11);
			goto fail;
		}
		message[msz] = '\0';
	}
	git_signature *sig = NULL;
	if (git_signature_new(&sig, aname, aemail, (git_time_t) when, (int)tzoff) != 0) {
		free(message);
		bare_put_uint(writer, 19);
		goto fail;
	}
	git_tree *tree = NULL;
	if (git_tree_lookup(&tree, repo, &tree_oid) != 0) {
		git_signature_free(sig);
		free(message);
		bare_put_uint(writer, 19);
		goto fail;
	}
	git_oid out;
	int rc = git_commit_create(&out, repo, NULL, sig, sig, NULL, message, tree,
				   (int)pcnt, (const git_commit **)parents);
	git_tree_free(tree);
	git_signature_free(sig);
	free(message);
	if (rc != 0) {
		bare_put_uint(writer, 19);
		goto fail;
	}
	bare_put_uint(writer, 0);
	bare_put_data(writer, out.id, GIT_OID_RAWSZ);
	if (parents) {
		for (uint64_t i = 0; i < pcnt; i++)
			if (parents[i])
				git_commit_free(parents[i]);
		free(parents);
	}
	return 0;
 fail:
	if (parents) {
		for (uint64_t i = 0; i < pcnt; i++)
			if (parents[i])
				git_commit_free(parents[i]);
		free(parents);
	}
	return -1;
}

int cmd_update_ref(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char refname[4096] = { 0 };
	char commithex[64] = { 0 };
	if (bare_get_data(reader, (uint8_t *) refname, sizeof(refname) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	if (bare_get_data(reader, (uint8_t *) commithex, sizeof(commithex) - 1)
	    != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid oid;
	if (git_oid_fromstr(&oid, commithex) != 0) {
		bare_put_uint(writer, 18);
		return -1;
	}
	git_reference *out = NULL;
	int rc = git_reference_create(&out, repo, refname, &oid, 1, NULL);
	if (rc != 0) {
		bare_put_uint(writer, 18);
		return -1;
	}
	git_reference_free(out);
	bare_put_uint(writer, 0);
	return 0;
}

int cmd_commit_info(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char hex[64] = { 0 };
	if (bare_get_data(reader, (uint8_t *) hex, sizeof(hex) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid oid;
	if (git_oid_fromstr(&oid, hex) != 0) {
		bare_put_uint(writer, 14);
		return -1;
	}
	git_commit *commit = NULL;
	if (git_commit_lookup(&commit, repo, &oid) != 0) {
		bare_put_uint(writer, 14);
		return -1;
	}

	const git_signature *author = git_commit_author(commit);
	const git_signature *committer = git_commit_committer(commit);

	const char *aname = author && author->name ? author->name : "";
	const char *aemail = author && author->email ? author->email : "";
	git_time_t awhen = author ? author->when.time : 0;
	int aoffset = author ? author->when.offset : 0;

	const char *cname = committer && committer->name ? committer->name : "";
	const char *cemail = committer && committer->email ? committer->email : "";
	git_time_t cwhen = committer ? committer->when.time : 0;
	int coffset = committer ? committer->when.offset : 0;

	const char *message = git_commit_message(commit);
	if (!message) message = "";

	bare_put_uint(writer, 0);
	/* Commit ID */
	const git_oid *cid = git_commit_id(commit);
	bare_put_data(writer, cid->id, GIT_OID_RAWSZ);
	/* Author */
	bare_put_data(writer, (const uint8_t *)aname, strlen(aname));
	bare_put_data(writer, (const uint8_t *)aemail, strlen(aemail));
	bare_put_i64(writer, (int64_t)awhen);
	bare_put_i64(writer, (int64_t)aoffset);
	/* Committer */
	bare_put_data(writer, (const uint8_t *)cname, strlen(cname));
	bare_put_data(writer, (const uint8_t *)cemail, strlen(cemail));
	bare_put_i64(writer, (int64_t)cwhen);
	bare_put_i64(writer, (int64_t)coffset);
	/* Message */
	bare_put_data(writer, (const uint8_t *)message, strlen(message));
	/* Parents */
	uint32_t pcnt = git_commit_parentcount(commit);
	bare_put_uint(writer, (uint64_t)pcnt);
	for (uint32_t i = 0; i < pcnt; i++) {
		const git_commit *p = NULL;
		if (git_commit_parent((git_commit **)&p, commit, i) == 0 && p) {
			const git_oid *po = git_commit_id(p);
			bare_put_data(writer, po->id, GIT_OID_RAWSZ);
			git_commit_free((git_commit *)p);
		} else {
			uint8_t zero[GIT_OID_RAWSZ] = {0};
			bare_put_data(writer, zero, GIT_OID_RAWSZ);
		}
	}

	/* Structured diff */
	git_tree *tree = NULL;
	if (git_commit_tree(&tree, commit) != 0) {
		git_commit_free(commit);
		bare_put_uint(writer, 15);
		return -1;
	}
	git_diff *diff = NULL;
	if (pcnt == 0) {
		if (git_diff_tree_to_tree(&diff, repo, NULL, tree, NULL) != 0) {
			git_tree_free(tree);
			git_commit_free(commit);
			bare_put_uint(writer, 15);
			return -1;
		}
	} else {
		git_commit *parent = NULL;
		git_tree *ptree = NULL;
		if (git_commit_parent(&parent, commit, 0) != 0 || git_commit_tree(&ptree, parent) != 0) {
			if (parent) git_commit_free(parent);
			git_tree_free(tree);
			git_commit_free(commit);
			bare_put_uint(writer, 15);
			return -1;
		}
		if (git_diff_tree_to_tree(&diff, repo, ptree, tree, NULL) != 0) {
			git_tree_free(ptree);
			git_commit_free(parent);
			git_tree_free(tree);
			git_commit_free(commit);
			bare_put_uint(writer, 15);
			return -1;
		}
		git_tree_free(ptree);
		git_commit_free(parent);
	}

	size_t files = git_diff_num_deltas(diff);
	bare_put_uint(writer, (uint64_t)files);
    for (size_t i = 0; i < files; i++) {
		git_patch *patch = NULL;
		if (git_patch_from_diff(&patch, diff, i) != 0) {
			/* empty diff */
			bare_put_uint(writer, 0);
			bare_put_uint(writer, 0);
			bare_put_data(writer, (const uint8_t *)"", 0);
			bare_put_data(writer, (const uint8_t *)"", 0);
			bare_put_uint(writer, 0);
			continue;
		}
		const git_diff_delta *delta = git_patch_get_delta(patch);
		uint32_t from_mode = delta ? delta->old_file.mode : 0;
		uint32_t to_mode = delta ? delta->new_file.mode : 0;
		const char *from_path = (delta && delta->old_file.path) ? delta->old_file.path : "";
		const char *to_path = (delta && delta->new_file.path) ? delta->new_file.path : "";
		bare_put_uint(writer, (uint64_t)from_mode);
		bare_put_uint(writer, (uint64_t)to_mode);
		bare_put_data(writer, (const uint8_t *)from_path, strlen(from_path));
		bare_put_data(writer, (const uint8_t *)to_path, strlen(to_path));

		size_t hunks = git_patch_num_hunks(patch);
		uint64_t chunk_count = 0;
		for (size_t h = 0; h < hunks; h++) {
			const git_diff_hunk *hunk = NULL;
			size_t lines = 0;
			if (git_patch_get_hunk(&hunk, &lines, patch, h) != 0) continue;
			int prev = -2;
			for (size_t ln = 0; ln < lines; ln++) {
				const git_diff_line *line = NULL;
				if (git_patch_get_line_in_hunk(&line, patch, h, ln) != 0 || !line) continue;
				int op = 0;
				if (line->origin == '+') op = 1;
				else if (line->origin == '-') op = 2;
				else op = 0;
				if (op != prev) { chunk_count++; prev = op; }
			}
		}
		bare_put_uint(writer, chunk_count);
        for (size_t h = 0; h < hunks; h++) {
            const git_diff_hunk *hunk = NULL;
            size_t lines = 0;
            if (git_patch_get_hunk(&hunk, &lines, patch, h) != 0) continue;
            int prev = -2;
            struct {
                char *data;
                size_t len;
                size_t cap;
            } buf = {0};
            for (size_t ln = 0; ln < lines; ln++) {
                const git_diff_line *line = NULL;
                if (git_patch_get_line_in_hunk(&line, patch, h, ln) != 0 || !line) continue;
                int op = 0;
                if (line->origin == '+') op = 1;
                else if (line->origin == '-') op = 2;
                else op = 0;
                if (prev == -2) prev = op;
                if (op != prev) {
                    bare_put_uint(writer, (uint64_t)prev);
                    bare_put_data(writer, (const uint8_t *)buf.data, buf.len);
                    free(buf.data);
                    buf.data = NULL; buf.len = 0; buf.cap = 0;
                    prev = op;
                }
                if (line->content && line->content_len > 0) {
                    if (append_buf(&buf.data, &buf.len, &buf.cap, line->content, line->content_len) != 0) {
                        free(buf.data);
                        git_patch_free(patch);
                        git_diff_free(diff);
                        git_tree_free(tree);
                        git_commit_free(commit);
                        bare_put_uint(writer, 15);
                        return -1;
                    }
                }
            }
            if (prev != -2) {
                bare_put_uint(writer, (uint64_t)prev);
                bare_put_data(writer, (const uint8_t *)buf.data, buf.len);
                free(buf.data);
            }
        }
        git_patch_free(patch);
    }

	git_diff_free(diff);
	git_tree_free(tree);
	git_commit_free(commit);
	return 0;
}
