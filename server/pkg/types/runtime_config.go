package types

// RuntimeConfig is the service-facing representation of a configuration.
type RuntimeConfig struct {
	Name        string `json:"name"`
	Format      int    `json:"format"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Version     uint   `json:"version"`
	UpdatedAt   string `json:"updated_at"`
}
