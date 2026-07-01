package versioning

import (
	"bank/internal/api/misc"
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

func GetVersionedEndpoints() VersionedEndpoints {
	return VersionedEndpoints{
		VersionV1dot0: {
			http.MethodGet: {
				misc.EndpointPing: misc.PingV1dot0,
			},
		},
	}
}
