package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
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
}
