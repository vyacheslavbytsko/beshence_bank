package versioning

import (
	"bank/internal/api"
	authend "bank/internal/api/endpoints/auth"
	"bank/internal/api/endpoints/misc"
	"bank/internal/auth"
	"bank/internal/middleware"
	"net/http"
)

const (
	HeaderAPIVersion  = "X-Beshence-Bank-API-Version"
	VersionV1dot0     = "v1.0"
	DefaultAPIVersion = VersionV1dot0
)

var supportedVersions = []string{VersionV1dot0 /*, VersionV1dot1*/}

var versionIndex = map[string]int{
	VersionV1dot0: 0,
	// VersionV1dot1: 1,
}

func GetVersionedEndpoints(deps *api.Dependencies) VersionedEndpoints {
	return VersionedEndpoints{
		VersionV1dot0: {
			http.MethodGet: {
				"/ping": misc.PingV1dot0,
				// TODO: remove `auth.TokenTypeAccess`/`auth.TokenTypeRefresh` because manager can handle it... i guess..
				"/auth/me":      middleware.RequireAuth(deps.AccessJWTManager, auth.TokenTypeAccess, authend.MeV1dot0(deps)),
				"/auth/refresh": middleware.RequireAuth(deps.RefreshJWTManager, auth.TokenTypeRefresh, authend.RefreshV1dot0(deps)),
			},
			http.MethodPost: {
				"/auth/register": authend.RegisterV1dot0(deps),
				"/auth/login":    authend.LoginV1dot0(deps),
			},
		},
	}
}
