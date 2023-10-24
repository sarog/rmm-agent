package config

type IAgentConfig interface {
	setConfig(config *AgentConfig)
	getConfig() *AgentConfig
}

type AgentConfig struct {
	AgentID  string // Username (as ULID)
	AgentPK  int    // Primary Key on server?
	BaseURL  string // Server URL
	ApiURL   string // NATS
	ApiPort  int    // NATS Port (4222)
	Token    string // Authorization token
	Cert     string // Root Certificate
	Arch     string // "x86_64", "x86"
	Debug    bool
	Hostname string
	Version  string
	Headers  map[string]string
}
