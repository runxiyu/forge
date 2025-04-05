package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
	goSSH "golang.org/x/crypto/ssh"
)

type server struct {
	config Config

	// database serves as the primary database handle for this entire application.
	// Transactions or single reads may be used from it. A [pgxpool.Pool] is
	// necessary to safely use pgx concurrently; pgx.Conn, etc. are insufficient.
	database *pgxpool.Pool

	sourceHandler http.Handler
	staticHandler http.Handler

	ircSendBuffered   chan string
	ircSendDirectChan chan errorBack[string]

	// globalData is passed as "global" when rendering HTML templates.
	globalData map[string]any

	serverPubkeyString string
	serverPubkeyFP     string
	serverPubkey       goSSH.PublicKey

	// packPasses contains hook cookies mapped to their packPass.
	packPasses cmap.Map[string, packPass]
}
