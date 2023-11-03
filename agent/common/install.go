package common

import "time"

type InstallInfo struct {
	Headers     map[string]string
	ServerURL   string        // JSON endpoint URL
	ApiURL      string        // RPC endpoint (NATS) URL
	ClientID    int           // Client ID
	SiteID      int           // Client Site ID
	Description string        // Defaults to hostname
	Token       string        // Authorization token (password)
	RootCert    string        // Trusted Root Certificate
	Timeout     time.Duration // Installation timeout
	Silent      bool          // Silent installation
	// AgentType   string // Workstation, Server
}
