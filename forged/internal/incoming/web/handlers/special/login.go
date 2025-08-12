package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/argon2id"
	"go.lindenii.runxiyu.org/forge/forged/internal/common/misc"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/templates"
	wtypes "go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

type LoginHTTP struct {
	r            templates.Renderer
	cookieExpiry int
}

func NewLoginHTTP(r templates.Renderer, cookieExpiry int) *LoginHTTP {
	return &LoginHTTP{
		r:            r,
		cookieExpiry: cookieExpiry,
	}
}

func (h *LoginHTTP) Login(w http.ResponseWriter, r *http.Request, _ wtypes.Vars) {
	renderLoginPage := func(loginError string) bool {
		err := h.r.Render(w, "login", struct {
			BaseData   *wtypes.BaseData
			LoginError string
		}{
			BaseData:   wtypes.Base(r),
			LoginError: loginError,
		})
		if err != nil {
			log.Println("failed to render login page", "error", err)
			http.Error(w, "Failed to render login page", http.StatusInternalServerError)
			return true
		}
		return false
	}

	if r.Method == http.MethodGet {
		renderLoginPage("")
		return
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	userCreds, err := wtypes.Base(r).Global.Queries.GetUserCreds(r.Context(), &username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			renderLoginPage("User not found")
			return
		}
		log.Println("failed to get user credentials", "error", err)
		http.Error(w, "Failed to get user credentials", http.StatusInternalServerError)
		return
	}

	if userCreds.PasswordHash == "" {
		renderLoginPage("No password set for this user")
		return
	}

	passwordMatches, err := argon2id.ComparePasswordAndHash(password, userCreds.PasswordHash)
	if err != nil {
		log.Println("failed to compare password and hash", "error", err)
		http.Error(w, "Failed to verify password", http.StatusInternalServerError)
		return
	}

	if !passwordMatches {
		renderLoginPage("Invalid password")
		return
	}

	cookieValue := rand.Text()

	now := time.Now()
	expiry := now.Add(time.Duration(h.cookieExpiry) * time.Second)

	cookie := &http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   false, // TODO
		Expires:  expiry,
		Path:     "/",
	} //exhaustruct:ignore

	http.SetCookie(w, cookie)

	tokenHash := sha256.Sum256(misc.StringToBytes(cookieValue))

	err = wtypes.Base(r).Global.Queries.InsertSession(r.Context(), queries.InsertSessionParams{
		UserID:    userCreds.ID,
		TokenHash: tokenHash[:],
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiry,
			Valid: true,
		},
	})
	if err != nil {
		log.Println("failed to insert session", "error", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
