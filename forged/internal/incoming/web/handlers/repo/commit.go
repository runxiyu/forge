package repo

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
	"go.lindenii.runxiyu.org/forge/forged/internal/ipc/git2c"
)

type commitPerson struct {
	Name  string
	Email string
	When  time.Time
}

type commitObject struct {
	Hash      string
	Message   string
	Author    commitPerson
	Committer commitPerson
}

type usableChunk struct {
	Operation int
	Content   string
}

type diffFileMeta struct {
	Hash string
	Mode string
	Path string
}

type usableFilePatch struct {
	From   diffFileMeta
	To     diffFileMeta
	Chunks []usableChunk
}

func shortHash(s string) string {
	if s == "" {
		return ""
	}
	b := sha1.Sum([]byte(s))
	return hex.EncodeToString(b[:8])
}

func parseUnifiedPatch(p string) []usableFilePatch {
	lines := strings.Split(p, "\n")
	patches := []usableFilePatch{}
	var cur *usableFilePatch
	flush := func() {
		if cur != nil {
			patches = append(patches, *cur)
			cur = nil
		}
	}
	appendChunk := func(op int, buf *[]string) {
		if len(*buf) == 0 || cur == nil {
			return
		}
		content := strings.Join(*buf, "\n")
		*buf = (*buf)[:0]
		cur.Chunks = append(cur.Chunks, usableChunk{Operation: op, Content: content})
	}
	var bufSame, bufAdd, bufDel []string

	for _, ln := range lines {
		if strings.HasPrefix(ln, "diff --git ") {
			appendChunk(0, &bufSame)
			appendChunk(1, &bufAdd)
			appendChunk(2, &bufDel)
			flush()
			parts := strings.SplitN(strings.TrimPrefix(ln, "diff --git "), " ", 2)
			from := strings.TrimPrefix(strings.TrimSpace(parts[0]), "a/")
			to := from
			if len(parts) > 1 {
				to = strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(parts[1], "b/")), "b/")
			}
			cur = &usableFilePatch{
				From: diffFileMeta{Path: from, Hash: shortHash(from)},
				To:   diffFileMeta{Path: to, Hash: shortHash(to)},
			}
			continue
		}
		if cur == nil {
			continue
		}
		switch {
		case strings.HasPrefix(ln, "+"):
			appendChunk(0, &bufSame)
			appendChunk(2, &bufDel)
			bufAdd = append(bufAdd, ln)
		case strings.HasPrefix(ln, "-"):
			appendChunk(0, &bufSame)
			appendChunk(1, &bufAdd)
			bufDel = append(bufDel, ln)
		default:
			appendChunk(1, &bufAdd)
			appendChunk(2, &bufDel)
			bufSame = append(bufSame, ln)
		}
	}
	if cur != nil {
		appendChunk(0, &bufSame)
		appendChunk(1, &bufAdd)
		appendChunk(2, &bufDel)
		flush()
	}
	return patches
}

func (h *HTTP) Commit(w http.ResponseWriter, r *http.Request, v wtypes.Vars) {
	base := wtypes.Base(r)
	repoName := v["repo"]
	commitSpec := v["commit"]
	wantPatch := strings.HasSuffix(commitSpec, ".patch")
	commitSpec = strings.TrimSuffix(commitSpec, ".patch")

	var userID int64
	if base.UserID != "" {
		_, _ = fmt.Sscan(base.UserID, &userID)
	}
	grp, err := base.Global.Queries.GetGroupByPath(r.Context(), queries.GetGroupByPathParams{Column1: base.GroupPath, UserID: userID})
	if err != nil {
		slog.Error("get group by path", "error", err)
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	repoRow, err := base.Global.Queries.GetRepoByGroupAndName(r.Context(), queries.GetRepoByGroupAndNameParams{GroupID: grp.ID, Name: repoName})
	if err != nil {
		slog.Error("get repo by name", "error", err)
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	repoPath := filepath.Join(base.Global.Config.Git.RepoDir, fmt.Sprintf("%d.git", repoRow.ID))
	client, err := git2c.NewClient(r.Context(), base.Global.Config.Git.Socket)
	if err != nil {
		slog.Error("git2d connect failed", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer func() { _ = client.Close() }()

	resolved := commitSpec
	if len(commitSpec) < 40 {
		if list, lerr := client.Log(repoPath, commitSpec, 1); lerr == nil && len(list) > 0 {
			resolved = list[0].Hash
		}
	}
	if !wantPatch && resolved != "" && resolved != commitSpec {
		u := *r.URL
		basePath := strings.TrimSuffix(u.EscapedPath(), commitSpec)
		u.Path = basePath + resolved
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
		return
	}

	if wantPatch {
		patchStr, perr := client.FormatPatch(repoPath, resolved)
		if perr != nil {
			slog.Error("format patch failed", "error", perr)
			http.Error(w, "Failed to format patch", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(patchStr))
		return
	}

	info, derr := client.CommitInfo(repoPath, resolved)
	if derr != nil {
		slog.Error("commit info failed", "error", derr)
		http.Error(w, "Failed to get commit info", http.StatusInternalServerError)
		return
	}

	toTime := func(sec, minoff int64) time.Time {
		loc := time.FixedZone("", int(minoff*60))
		return time.Unix(sec, 0).In(loc)
	}
	co := commitObject{
		Hash:      info.Hash,
		Message:   info.Message,
		Author:    commitPerson{Name: info.AuthorName, Email: info.AuthorEmail, When: toTime(info.AuthorWhen, info.AuthorTZMin)},
		Committer: commitPerson{Name: info.CommitterName, Email: info.CommitterEmail, When: toTime(info.CommitterWhen, info.CommitterTZMin)},
	}

	toUsable := func(files []git2c.FileDiff) []usableFilePatch {
		out := make([]usableFilePatch, 0, len(files))
		for _, f := range files {
			u := usableFilePatch{
				From: diffFileMeta{Path: f.FromPath, Mode: fmt.Sprintf("%06o", f.FromMode), Hash: shortHash(f.FromPath)},
				To:   diffFileMeta{Path: f.ToPath, Mode: fmt.Sprintf("%06o", f.ToMode), Hash: shortHash(f.ToPath)},
			}
			for _, ch := range f.Chunks {
				u.Chunks = append(u.Chunks, usableChunk{Operation: int(ch.Op), Content: ch.Content})
			}
			out = append(out, u)
		}
		return out
	}
	filePatches := toUsable(info.Files)
	parentHex := ""
	if len(info.Parents) > 0 {
		parentHex = info.Parents[0]
	}

	repoURLRoot := "/" + misc.SegmentsToURL(base.GroupPath) + "/-/repos/" + url.PathEscape(repoRow.Name) + "/"
	data := map[string]any{
		"BaseData":           base,
		"group_path":         base.GroupPath,
		"repo_name":          repoRow.Name,
		"repo_description":   repoRow.Description,
		"repo_url_root":      repoURLRoot,
		"commit_object":      co,
		"commit_id":          co.Hash,
		"parent_commit_hash": parentHex,
		"file_patches":       filePatches,
		"global": map[string]any{
			"forge_title": base.Global.ForgeTitle,
		},
	}
	if err := h.r.Render(w, "repo_commit", data); err != nil {
		slog.Error("render repo commit", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
