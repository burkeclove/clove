package responses

type CreateSigV4Response struct {
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	SessionToken string `json:"session_token"`
	ExpiresAt    string `json:"expires_at"`
}
