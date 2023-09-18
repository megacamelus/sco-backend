// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/sco1237896/sco-backend/app/services/sco-api/handlers/v1/checkgrp"
	"github.com/sco1237896/sco-backend/business/web/auth"
	"github.com/sco1237896/sco-backend/foundation/logger"
	"github.com/sco1237896/sco-backend/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build string
	Log   *logger.Logger
	Auth  *auth.Auth
	DB    *sqlx.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// -------------------------------------------------------------------------

	cgh := checkgrp.New(cfg.Build)

	app.HandleNoMiddleware(http.MethodGet, version, "/readiness", cgh.Readiness)
	app.HandleNoMiddleware(http.MethodGet, version, "/liveness", cgh.Liveness)

}
