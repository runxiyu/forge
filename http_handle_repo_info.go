package main

import (
	"net/http"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

func handle_repo_info(w http.ResponseWriter, r *http.Request, params map[string]any) (err error) {
	group_name, repo_name := params["group_name"].(string), params["repo_name"].(string)
	var repo_path string
	err = database.QueryRow(r.Context(), "SELECT r.filesystem_path FROM repos r JOIN groups g ON r.group_id = g.id WHERE g.name = $1 AND r.name = $2;", group_name, repo_name).Scan(&repo_path)
	if err != nil {
		return err
	}
	endpoint, err := transport.NewEndpoint("/")
	if err != nil {
		return err
	}
	billy_fs := osfs.New(repo_path)
	fs_loader := server.NewFilesystemLoader(billy_fs)
	transport := server.NewServer(fs_loader)
	upload_pack_session, err := transport.NewUploadPackSession(endpoint, nil)
	if err != nil {
		return err
	}
	advertised_references, err := upload_pack_session.AdvertisedReferencesContext(r.Context())
	if err != nil {
		return err
	}
	advertised_references.Prefix = [][]byte{[]byte("# service=git-upload-pack"), pktline.Flush}
	w.Header().Set("content-type", "application/x-git-upload-pack-advertisement")
	err = advertised_references.Encode(w)
	if err != nil {
		return err
	}
	return nil
}
