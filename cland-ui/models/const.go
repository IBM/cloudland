package models

var routeInfo = map[string]string{
	"volume":        `/volume/v3`,
	"compute":       `/compute/v2.1`,
	"image":         `/image`,
	"network":       `/v2.0/networks`,
	"identity":      `/identity`,
	"subnet":        `/v2.0/subnets`,
	"identityToken": "/auth/tokens",
	"flavor":        "/compute/v2.1/flavors",
}

const TOKEN_HEADER_KEY = `X-Subject-Token`
