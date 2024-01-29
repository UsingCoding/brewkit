package config

type Config struct {
	Secrets      []Secret `json:"secrets"`
	Entitlements []string `json:"entitlements"`
}

type Secret struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}
