package utils

import "time"

const (
	AuthCSRFCookie   = "auth-csrf"
	AuthAccessCookie = "auth-access"
	TokenLength      = 32
	TokenExpiryTime  = 5 * time.Minute
)
