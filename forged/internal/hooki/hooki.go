package hooki

import (
	"go.lindenii.runxiyu.org/forge/forged/internal/cmap"
	"github.com/gliderlabs/ssh"
)

type Pool cmap.Map[string, hookinfo]

type hookinfo struct {
	session      ssh.Session
	pubkey       string
	directAccess bool
	repoPath     string
	userID       int
	userType     string
	repoID       int
	groupPath    []string
	repoName     string
	contribReq   string
}
