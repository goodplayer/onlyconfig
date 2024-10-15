package config

type WebManagerConfig struct {
	JwtSecrets [][]byte `json:"jwt_secrets"`
}
