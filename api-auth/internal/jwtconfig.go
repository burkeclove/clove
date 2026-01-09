package internal

import (
	"time"
)

type Config struct {
	Issuer   string        // "iss"
	Audience string        // "aud" 
	AccessTTL time.Duration // token lifetime

	Leeway time.Duration // allowed clock skew
	KeyID string
}
