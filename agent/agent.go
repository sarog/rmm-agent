package agent

import (
	"github.com/go-resty/resty/v2"
	"github.com/sarog/rmmagent/shared"
	"github.com/sarog/trmm-shared"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

const (
	AGENT_FOLDER        = "RMMAgent"
	AGENT_NAME_LONG     = "RMM Agent"
	AGENT_TEMP_DIR      = "rmm"
	NATS_RMM_IDENTIFIER = "ACMERMM"
	NATS_DEFAULT_PORT   = 4222
	RMM_SEARCH_PREFIX   = "acmermm*"
)

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

type Installer struct {
	// Headers     map[string]string
	ServerURL   string // dupe?
	ClientID    int
	SiteID      int
	Description string
	// AgentType   string // Workstation, Server
	Token   string // dupe?
	Cert    string // dupe?
	Timeout time.Duration
	Silent  bool
}

type InfoCollector interface {
	PublicIP() string
	TotalRAM() float64
	BootTime() int64
	GetInstalledSoftware() []shared.SoftwareList
	OSInfo() (plat, osFullName string)
	SysInfo()
	GetDisksNATS() []trmm.Disk
	LoggedOnUser() string
	GetCPULoadAvg() int
}

type TaskChecker interface {
	ScriptCheck(data shared.Check, r *resty.Client)
	DiskCheck(data shared.Check, r *resty.Client)
	CPULoadCheck(data shared.Check, r *resty.Client)
	MemCheck(data shared.Check, r *resty.Client)
	PingCheck(data shared.Check, r *resty.Client)
}

// todo
type TaskRunner interface {
}

// todo
type ServiceManager interface {
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
}

type Interface interface {
	New(logger *logrus.Logger, version string) (Agent, error) // todo: move this?

	// New(i Interface, c *AgentConfig) (Agent, error)

	// Setup
	Install(i *Installer)
	AgentUpdate(url, inno, version string)
	AgentUninstall()
	UninstallCleanup()

	// Service management
	RunAgentService()
	// Deprecated replace with combined Service
	RunRPCService()

	Hostname() string
	ShowStatus(version string)

	RunTask(int)
	RunChecks(force bool) error
	RunScript(code string, shell string, args []string, timeout int) (stdout, stderr string, exitcode int, e error)
	CheckIn(mode string)
	CreateInternalTask(name, args, repeat string, start int) (bool, error)
	CheckRunner()
	GetCheckInterval() (int, error)

	// Transmit
	SendSoftware()
	SyncInfo()

	RecoverAgent()

	// Windows-specific:
	// InstallUpdates(guids []string)
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
	// GetServiceDetail(name string) shared.WindowsService
	// GetServicesNATS() []trmm.WindowsService
	// GetServices() []shared.WindowsService
	// CreateSchedTask(st windows.SchedTask) (bool, error)
}

type Agent struct {
	Interface
	AgentConfig
	InfoCollector
	Logger  *logrus.Logger
	RClient *resty.Client
	// Logger *logrus.Logger
}

// GenerateAgentID creates and returns a unique agent ID
// todo: what about UUIDs?
func GenerateAgentID() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func ShowStatus(a Agent, version string) {
	a.ShowStatus(version)
}
