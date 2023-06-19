package utils

import "time"

const (
	AuthCSRFCookie  = "auth-csrf"
	AuthIdCookie    = "auth-id"
	TokenLength     = 32
	TokenExpiryTime = 5 * time.Minute
)
