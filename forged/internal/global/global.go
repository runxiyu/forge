package global

import (
	"go.lindenii.runxiyu.org/forge/forged/internal/config"
	"go.lindenii.runxiyu.org/forge/forged/internal/database"
	"go.lindenii.runxiyu.org/forge/forged/internal/database/queries"
)

type Global struct {
	ForgeTitle     string // should be removed since it's in Config
	ForgeVersion   string
	SSHPubkey      string
	SSHFingerprint string

	Config  *config.Config
	Queries *queries.Queries
	DB      *database.Database
}
