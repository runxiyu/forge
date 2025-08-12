/*-
 * SPDX-License-Identifier: AGPL-3.0-only
 * SPDX-FileCopyrightText: Copyright (c) 2025 Runxi Yu <https://runxiyu.org>
 */

#include "x.h"

static int diff_stats_to_string(git_diff *diff, git_buf *out)
{
	git_diff_stats *stats = NULL;
	if (git_diff_get_stats(&stats, diff) != 0) {
		return -1;
	}
	int rc = git_diff_stats_to_buf(out, stats, GIT_DIFF_STATS_FULL, 80);
	git_diff_stats_free(stats);
	return rc;
}

static void split_message(const char *message, char **title_out, char **body_out)
{
	*title_out = NULL;
	*body_out = NULL;
	if (!message)
		return;
	const char *nl = strchr(message, '\n');
	if (!nl) {
		*title_out = strdup(message);
		*body_out = strdup("");
		return;
	}
	size_t title_len = (size_t)(nl - message);
	*title_out = (char *)malloc(title_len + 1);
	if (*title_out) {
		memcpy(*title_out, message, title_len);
		(*title_out)[title_len] = '\0';
	}
	const char *rest = nl + 1;
	if (*rest == '\n')
		rest++;
	*body_out = strdup(rest);
}

int cmd_format_patch(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
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

	git_diff *diff = NULL;
	if (git_commit_parentcount(commit) == 0) {
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
			if (parent)
				git_commit_free(parent);
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

	git_buf stats = { 0 };
	if (diff_stats_to_string(diff, &stats) != 0) {
		git_diff_free(diff);
		git_tree_free(tree);
		git_commit_free(commit);
		bare_put_uint(writer, 15);
		return -1;
	}

	git_buf patch = { 0 };
	if (git_diff_to_buf(&patch, diff, GIT_DIFF_FORMAT_PATCH) != 0) {
		git_buf_dispose(&stats);
		git_diff_free(diff);
		git_tree_free(tree);
		git_commit_free(commit);
		bare_put_uint(writer, 15);
		return -1;
	}

	const git_signature *author = git_commit_author(commit);
	char *title = NULL, *body = NULL;
	split_message(git_commit_message(commit), &title, &body);

	char header[2048];
	char timebuf[64];
	{
		time_t t = git_commit_time(commit);
		struct tm *tm = localtime(&t);
		if (tm)
			strftime(timebuf, sizeof(timebuf), "%a, %d %b %Y %H:%M:%S %z", tm);
		else
			strcpy(timebuf, "unknown");
	}
	snprintf(header, sizeof(header), "From %s Mon Sep 17 00:00:00 2001\nFrom: %s <%s>\nDate: %s\nSubject: [PATCH] %s\n\n", git_oid_tostr_s(&oid), author && author->name ? author->name : "", author && author->email ? author->email : "", timebuf, title ? title : "");

	const char *trailer = "\n-- \n2.48.1\n";
	size_t header_len = strlen(header);
	size_t body_len = body ? strlen(body) : 0;
	size_t trailer_len = strlen(trailer);
	size_t total = header_len + body_len + (body_len ? 1 : 0) + 4 + stats.size + 1 + patch.size + trailer_len;

	uint8_t *buf = (uint8_t *) malloc(total);
	if (!buf) {
		free(title);
		free(body);
		git_buf_dispose(&patch);
		git_buf_dispose(&stats);
		git_diff_free(diff);
		git_tree_free(tree);
		git_commit_free(commit);
		bare_put_uint(writer, 15);
		return -1;
	}
	size_t off = 0;
	memcpy(buf + off, header, header_len);
	off += header_len;
	if (body_len) {
		memcpy(buf + off, body, body_len);
		off += body_len;
		buf[off++] = '\n';
	}
	memcpy(buf + off, "---\n", 4);
	off += 4;
	memcpy(buf + off, stats.ptr, stats.size);
	off += stats.size;
	buf[off++] = '\n';
	memcpy(buf + off, patch.ptr, patch.size);
	off += patch.size;
	memcpy(buf + off, trailer, trailer_len);
	off += trailer_len;

	bare_put_uint(writer, 0);
	bare_put_data(writer, buf, off);

	free(buf);
	free(title);
	free(body);
	git_buf_dispose(&patch);
	git_buf_dispose(&stats);
	git_diff_free(diff);
	git_tree_free(tree);
	git_commit_free(commit);
	return 0;
}

int cmd_merge_base(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char hex1[64] = { 0 };
	char hex2[64] = { 0 };
	if (bare_get_data(reader, (uint8_t *) hex1, sizeof(hex1) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	if (bare_get_data(reader, (uint8_t *) hex2, sizeof(hex2) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	git_oid a, b, out;
	if (git_oid_fromstr(&a, hex1) != 0 || git_oid_fromstr(&b, hex2) != 0) {
		bare_put_uint(writer, 17);
		return -1;
	}
	int rc = git_merge_base(&out, repo, &a, &b);
	if (rc == GIT_ENOTFOUND) {
		bare_put_uint(writer, 16);
		return -1;
	}
	if (rc != 0) {
		bare_put_uint(writer, 17);
		return -1;
	}
	bare_put_uint(writer, 0);
	bare_put_data(writer, out.id, GIT_OID_RAWSZ);
	return 0;
}

int cmd_log(git_repository *repo, struct bare_reader *reader, struct bare_writer *writer)
{
	char spec[4096] = { 0 };
	uint64_t limit = 0;
	if (bare_get_data(reader, (uint8_t *) spec, sizeof(spec) - 1) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}
	if (bare_get_uint(reader, &limit) != BARE_ERROR_NONE) {
		bare_put_uint(writer, 11);
		return -1;
	}

	git_object *obj = NULL;
	if (spec[0] == '\0')
		strcpy(spec, "HEAD");
	if (git_revparse_single(&obj, repo, spec) != 0) {
		bare_put_uint(writer, 4);
		return -1;
	}
	git_commit *start = (git_commit *) obj;

	git_revwalk *walk = NULL;
	if (git_revwalk_new(&walk, repo) != 0) {
		git_commit_free(start);
		bare_put_uint(writer, 9);
		return -1;
	}
	git_revwalk_sorting(walk, GIT_SORT_TIME);
	git_revwalk_push(walk, git_commit_id(start));
	git_commit_free(start);

	bare_put_uint(writer, 0);
	git_oid oid;
	uint64_t count = 0;
	while ((limit == 0 || count < limit)
	       && git_revwalk_next(&oid, walk) == 0) {
		git_commit *c = NULL;
		if (git_commit_lookup(&c, repo, &oid) != 0)
			break;
		const char *msg = git_commit_summary(c);
		const git_signature *author = git_commit_author(c);
		time_t t = git_commit_time(c);
		char timebuf[64];
		struct tm *tm = localtime(&t);
		if (tm)
			strftime(timebuf, sizeof(timebuf), "%Y-%m-%d %H:%M:%S", tm);
		else
			strcpy(timebuf, "unknown");

		bare_put_data(writer, oid.id, GIT_OID_RAWSZ);
		bare_put_data(writer, (const uint8_t *)(msg ? msg : ""), msg ? strlen(msg) : 0);
		bare_put_data(writer, (const uint8_t *)(author && author->name ? author->name : ""), author && author->name ? strlen(author->name) : 0);
		bare_put_data(writer, (const uint8_t *)(author && author->email ? author->email : ""), author && author->email ? strlen(author->email) : 0);
		bare_put_data(writer, (const uint8_t *)timebuf, strlen(timebuf));
		git_commit_free(c);
		count++;
	}
	git_revwalk_free(walk);
	return 0;
}
