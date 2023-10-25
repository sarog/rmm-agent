package shared

// Deprecated (was "Checkin")
type AgentHeader struct {
	Func    string `json:"func"`
	AgentId string `json:"agent_id"`
	Version string `json:"version"`
}

type RecoveryAction struct {
	Mode     string `json:"mode"` // command, rpc
	ShellCMD string `json:"shellcmd"`
}

type AgentNeedsReboot struct {
	AgentID     string `json:"agent_id"`
	NeedsReboot bool   `json:"needs_reboot"`
}

// Storage holds physical disk info
// Deprecated
type Storage struct {
	Device  string  `json:"device"`
	Fstype  string  `json:"fstype"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Percent float64 `json:"percent"`
}

type AssignedTask struct {
	TaskPK  int  `json:"id"`
	Enabled bool `json:"enabled"`
}

type Script struct {
	Shell string `json:"shell"`
	Code  string `json:"code"`
}

type CheckInfo struct {
	AgentPK  int `json:"agent"`
	Interval int `json:"check_interval"`
}

type Check struct {
	Script         Script         `json:"script"`
	AssignedTasks  []AssignedTask `json:"assigned_tasks"`
	CheckPK        int            `json:"id"`
	CheckType      string         `json:"check_type"`
	Storage        string         `json:"storage"`
	IP             string         `json:"ip"`
	ScriptArgs     []string       `json:"script_args"`
	Timeout        int            `json:"timeout"`
	ServiceName    string         `json:"svc_name"`
	LogName        string         `json:"log_name"`
	EventID        int            `json:"event_id"`
	SearchLastDays int            `json:"search_last_days"`
	Status         string         `json:"status"`
	// Threshold        int            `json:"threshold"`
	// PassStartPending bool           `json:"pass_if_start_pending"`
	// PassNotExist     bool           `json:"pass_if_svc_not_exist"`
	// RestartIfStopped bool           `json:"restart_if_stopped"`
	// EventIDWildcard  bool           `json:"event_id_is_wildcard"`
	// EventType        string         `json:"event_type"`
	// EventSource      string         `json:"event_source"`
	// EventMessage     string         `json:"event_message"`
	// FailWhen         string         `json:"fail_when"`
}

type AllChecks struct {
	CheckInfo
	Checks []Check
}

type AutomatedTask struct {
	ID         int      `json:"id"`
	TaskScript Script   `json:"script"`
	Timeout    int      `json:"timeout"`
	Enabled    bool     `json:"enabled"`
	Args       []string `json:"script_args"`
}

type EventLogMsg struct {
	Source    string `json:"source"`
	EventType string `json:"eventType"`
	EventID   uint32 `json:"eventID"`
	Message   string `json:"message"`
	Time      string `json:"time"`
	// Deprecated
	UID int `json:"uid"` // for vue
}

type SoftwareList struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	InstallDate string `json:"install_date"`
	Size        string `json:"size"`
	Source      string `json:"source"`
	Location    string `json:"location"`
	Uninstall   string `json:"uninstall"`
}

type CheckInSW struct {
	AgentHeader
	InstalledSW []SoftwareList `json:"software"`
}

// Deprecated
type CheckInOS struct {
	AgentHeader
	Hostname     string  `json:"hostname"`
	OS           string  `json:"operating_system"`
	Platform     string  `json:"plat"`
	TotalRAM     float64 `json:"total_ram"`
	BootTime     int64   `json:"boot_time"`
	RebootNeeded bool    `json:"needs_reboot"`
	// todo: 2022-01-01: add: 'logged_in_username' ? natsapi/svc.go:77
}

// CheckInPublicIP
// Deprecated
type CheckInPublicIP struct {
	AgentHeader
	PublicIP string `json:"public_ip"`
}

// Deprecated
type CheckInDisk struct {
	AgentHeader
	Drives []Storage `json:"drives"`
}

// Deprecated
type CheckInLoggedUser struct {
	AgentHeader
	Username string `json:"logged_in_username"`
}
