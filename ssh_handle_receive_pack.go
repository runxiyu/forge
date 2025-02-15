package main

import (
	glider_ssh "github.com/gliderlabs/ssh"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	transport_server "github.com/go-git/go-git/v5/plumbing/transport/server"
)

func ssh_handle_receive_pack(session glider_ssh.Session, repo_identifier string) (err error) {
	repo_path, err := get_repo_path_from_ssh_path(session.Context(), repo_identifier)
	if err != nil {
		return err
	}
	endpoint, err := transport.NewEndpoint("/")
	if err != nil {
		return err
	}
	billy_fs := osfs.New(repo_path)
	fs_loader := transport_server.NewFilesystemLoader(billy_fs)
	transport := transport_server.NewServer(fs_loader)
	receive_pack_session, err := transport.NewReceivePackSession(endpoint, nil)
	if err != nil {
		return err
	}
	advertised_references, err := receive_pack_session.AdvertisedReferencesContext(session.Context())
	if err != nil {
		return err
	}
	err = advertised_references.Encode(session)
	if err != nil {
		return err
	}
	reference_update_request := packp.NewReferenceUpdateRequest()
	err = reference_update_request.Decode(session)
	if err != nil {
		return err
	}
	report_status, err := receive_pack_session.ReceivePack(session.Context(), reference_update_request)
	if err != nil {
		return err
	}
	err = report_status.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
