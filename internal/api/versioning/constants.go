package versioning

import (
	"bank/internal/api"
	"bank/internal/api/endpoints/auth"
	"bank/internal/api/endpoints/misc"
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
			},
			http.MethodPost: {
				"/auth/register": auth.RegisterV1dot0(deps),
				"/auth/login":    auth.LoginV1dot0(deps),
			},
		},
	}
}
