package common

const (
	AGENT_NAME_LONG   = "RMM Agent"
	AGENT_TEMP_DIR    = "rmm"
	NATS_DEFAULT_PORT = 4222
	// NATS_PROXY_PORT = 443
	// NATS_RMM_IDENTIFIER = "ACMERMM"

	TASK_PREFIX = "RMM_"
)

const (
	NATS_CMD_AGENT_UNINSTALL    = "uninstall"
	NATS_CMD_AGENT_UPDATE       = "agentupdate"
	NATS_CMD_CHOCO_INSTALL      = "installwithchoco"
	NATS_CMD_CPULOADAVG         = "cpuloadavg"
	NATS_CMD_EVENTLOG           = "eventlog"
	NATS_CMD_GETWINUPDATES      = "getwinupdates"
	NATS_CMD_INSTALL_CHOCO      = "installchoco"
	NATS_CMD_INSTALL_WINUPDATES = "installwinupdates"
	NATS_CMD_PING               = "ping"
	NATS_CMD_PROCS_KILL         = "killproc"
	NATS_CMD_PROCS_LIST         = "procs"
	NATS_CMD_PUBLICIP           = "publicip"
	NATS_CMD_RAWCMD             = "rawcmd"
	NATS_CMD_REBOOT_NEEDED      = "needsreboot"
	NATS_CMD_REBOOT_NOW         = "rebootnow"
	NATS_CMD_RECOVER            = "recover"
	NATS_CMD_RUNCHECKS          = "runchecks"
	NATS_CMD_SCRIPT_RUN         = "runscript"
	NATS_CMD_SCRIPT_RUN_FULL    = "runscriptfull"
	NATS_CMD_SOFTWARE_LIST      = "softwarelist"
	NATS_CMD_SYNC               = "sync"
	NATS_CMD_SYSINFO            = "sysinfo"
	NATS_CMD_TASK_ADD           = "schedtask"
	NATS_CMD_TASK_DEL           = "delschedtask"
	NATS_CMD_TASK_ENABLE        = "enableschedtask"
	NATS_CMD_TASK_LIST          = "listschedtasks"
	NATS_CMD_TASK_RUN           = "runtask"
	NATS_CMD_WINSERVICES        = "winservices"
	NATS_CMD_WINSVC_ACTION      = "winsvcaction"
	NATS_CMD_WINSVC_DETAIL      = "winsvcdetail"
	NATS_CMD_WINSVC_EDIT        = "editwinsvc"
	NATS_CMD_WMI                = "wmi"
)
