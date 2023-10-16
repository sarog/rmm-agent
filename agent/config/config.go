package config

type IAgentConfig interface {
	setConfig(config *AgentConfig)
	getConfig() *AgentConfig
}

type AgentConfig struct {
	AgentID  string
	AgentPK  string
	BaseURL  string // dupe?
	ApiURL   string // dupe?
	ApiPort  int
	Token    string
	PK       int
	Cert     string
	Arch     string // "x86_64", "x86"
	Debug    bool
	Hostname string
	Version  string
	Headers  map[string]string
}
