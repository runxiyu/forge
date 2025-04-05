package forge

import (
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.lindenii.runxiyu.org/lindenii-common/cmap"
	goSSH "golang.org/x/crypto/ssh"
)

type Server struct {
	Config Config

	// Database serves as the primary Database handle for this entire application.
	// Transactions or single reads may be used from it. A [pgxpool.Pool] is
	// necessary to safely use pgx concurrently; pgx.Conn, etc. are insufficient.
	Database *pgxpool.Pool

	SourceHandler http.Handler
	StaticHandler http.Handler

	IrcSendBuffered   chan string
	IrcSendDirectChan chan errorBack[string]

	// GlobalData is passed as "global" when rendering HTML templates.
	GlobalData map[string]any

	ServerPubkeyString string
	ServerPubkeyFP     string
	ServerPubkey       goSSH.PublicKey

	// PackPasses contains hook cookies mapped to their packPass.
	PackPasses cmap.Map[string, packPass]
}

func (s *Server) Setup() {
	s.SourceHandler = http.StripPrefix(
		"/-/source/",
		http.FileServer(http.FS(embeddedSourceFS)),
	)
	staticFS, err := fs.Sub(embeddedResourcesFS, "static")
	if err != nil {
		panic(err)
	}
	s.StaticHandler = http.StripPrefix("/-/static/", http.FileServer(http.FS(staticFS)))
	s.GlobalData = map[string]any{
		"server_public_key_string":      &s.ServerPubkeyString,
		"server_public_key_fingerprint": &s.ServerPubkeyFP,
		"forge_version":                 VERSION,
		// Some other ones are populated after config parsing
	}
}

func (s *Server) Run() {
	if err := s.deployHooks(); err != nil {
		slog.Error("deploying hooks", "error", err)
		os.Exit(1)
	}
	if err := loadTemplates(); err != nil {
		slog.Error("loading templates", "error", err)
		os.Exit(1)
	}
	if err := s.deployGit2D(); err != nil {
		slog.Error("deploying git2d", "error", err)
		os.Exit(1)
	}

	// Launch Git2D
	go func() {
		cmd := exec.Command(s.Config.Git.DaemonPath, s.Config.Git.Socket) //#nosec G204
		cmd.Stderr = log.Writer()
		cmd.Stdout = log.Writer()
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}()

	// UNIX socket listener for hooks
	{
		hooksListener, err := net.Listen("unix", s.Config.Hooks.Socket)
		if errors.Is(err, syscall.EADDRINUSE) {
			slog.Warn("removing existing socket", "path", s.Config.Hooks.Socket)
			if err = syscall.Unlink(s.Config.Hooks.Socket); err != nil {
				slog.Error("removing existing socket", "path", s.Config.Hooks.Socket, "error", err)
				os.Exit(1)
			}
			if hooksListener, err = net.Listen("unix", s.Config.Hooks.Socket); err != nil {
				slog.Error("listening hooks", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening hooks", "error", err)
			os.Exit(1)
		}
		slog.Info("listening hooks on unix", "path", s.Config.Hooks.Socket)
		go func() {
			if err = s.serveGitHooks(hooksListener); err != nil {
				slog.Error("serving hooks", "error", err)
				os.Exit(1)
			}
		}()
	}

	// UNIX socket listener for LMTP
	{
		lmtpListener, err := net.Listen("unix", s.Config.LMTP.Socket)
		if errors.Is(err, syscall.EADDRINUSE) {
			slog.Warn("removing existing socket", "path", s.Config.LMTP.Socket)
			if err = syscall.Unlink(s.Config.LMTP.Socket); err != nil {
				slog.Error("removing existing socket", "path", s.Config.LMTP.Socket, "error", err)
				os.Exit(1)
			}
			if lmtpListener, err = net.Listen("unix", s.Config.LMTP.Socket); err != nil {
				slog.Error("listening LMTP", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening LMTP", "error", err)
			os.Exit(1)
		}
		slog.Info("listening LMTP on unix", "path", s.Config.LMTP.Socket)
		go func() {
			if err = s.serveLMTP(lmtpListener); err != nil {
				slog.Error("serving LMTP", "error", err)
				os.Exit(1)
			}
		}()
	}

	// SSH listener
	{
		sshListener, err := net.Listen(s.Config.SSH.Net, s.Config.SSH.Addr)
		if errors.Is(err, syscall.EADDRINUSE) && s.Config.SSH.Net == "unix" {
			slog.Warn("removing existing socket", "path", s.Config.SSH.Addr)
			if err = syscall.Unlink(s.Config.SSH.Addr); err != nil {
				slog.Error("removing existing socket", "path", s.Config.SSH.Addr, "error", err)
				os.Exit(1)
			}
			if sshListener, err = net.Listen(s.Config.SSH.Net, s.Config.SSH.Addr); err != nil {
				slog.Error("listening SSH", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening SSH", "error", err)
			os.Exit(1)
		}
		slog.Info("listening SSH on", "net", s.Config.SSH.Net, "addr", s.Config.SSH.Addr)
		go func() {
			if err = s.serveSSH(sshListener); err != nil {
				slog.Error("serving SSH", "error", err)
				os.Exit(1)
			}
		}()
	}

	// HTTP listener
	{
		httpListener, err := net.Listen(s.Config.HTTP.Net, s.Config.HTTP.Addr)
		if errors.Is(err, syscall.EADDRINUSE) && s.Config.HTTP.Net == "unix" {
			slog.Warn("removing existing socket", "path", s.Config.HTTP.Addr)
			if err = syscall.Unlink(s.Config.HTTP.Addr); err != nil {
				slog.Error("removing existing socket", "path", s.Config.HTTP.Addr, "error", err)
				os.Exit(1)
			}
			if httpListener, err = net.Listen(s.Config.HTTP.Net, s.Config.HTTP.Addr); err != nil {
				slog.Error("listening HTTP", "error", err)
				os.Exit(1)
			}
		} else if err != nil {
			slog.Error("listening HTTP", "error", err)
			os.Exit(1)
		}
		server := http.Server{
			Handler:      s,
			ReadTimeout:  time.Duration(s.Config.HTTP.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(s.Config.HTTP.ReadTimeout) * time.Second,
			IdleTimeout:  time.Duration(s.Config.HTTP.ReadTimeout) * time.Second,
		} //exhaustruct:ignore
		slog.Info("listening HTTP on", "net", s.Config.HTTP.Net, "addr", s.Config.HTTP.Addr)
		go func() {
			if err = server.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("serving HTTP", "error", err)
				os.Exit(1)
			}
		}()
	}

	// IRC bot
	go s.ircBotLoop()

	select {}
}
