package web

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"go.lindenii.runxiyu.org/forge/forged/internal/incoming/web/types"
)

func userResolver(r *http.Request) (string, string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", "", nil
		}
		return "", "", err
	}

	tokenHash := sha256.Sum256([]byte(cookie.Value))

	session, err := types.Base(r).Global.Queries.GetUserFromSession(r.Context(), tokenHash[:])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}

	return fmt.Sprint(session.UserID), session.Username, nil
}
