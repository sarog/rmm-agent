package common

import "time"

type InstallInfo struct {
	Headers     map[string]string
	ServerURL   string // dupe?
	ApiURL      string // dupe? temp
	ClientID    int
	SiteID      int
	Description string
	// AgentType   string // Workstation, Server
	Token   string // dupe?
	Cert    string // dupe?
	Timeout time.Duration
	Silent  bool
}
