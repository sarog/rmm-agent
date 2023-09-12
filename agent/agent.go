package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sarog/rmmagent/shared"
	"github.com/sarog/trmm-shared"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	AGENT_FOLDER        = "RMMAgent"
	API_URL_SOFTWARE    = "/api/v3/software/"
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
	Logger   *logrus.Logger
	RClient  *resty.Client
}

type Installer struct {
	// Headers     map[string]string
	ServerURL   string // dupe?
	ClientID    int
	SiteID      int
	Description string
	// AgentType   string
	// Token        string
	// Cert         string
	Timeout time.Duration
	Silent  bool
}

type Agent interface {
	// InstallUpdates(guids []string)
	// New(logger *logrus.Logger, version string) *Agent
	PublicIP() string
	TotalRAM() float64
	BootTime() int64
	RunAgentService()
	CheckIn(mode string)
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
	// GetServiceDetail(name string) shared.WindowsService
	// GetServicesNATS() []trmm.WindowsService
	// GetServices() []shared.WindowsService
	// CreateSchedTask(st windows.SchedTask) (bool, error)
	// Install(i *windows.Installer)
	GetInstalledSoftware() []shared.SoftwareList
	CreateInternalTask(name, args, repeat string, start int) (bool, error)
	CheckRunner()
	GetCheckInterval() (int, error)
	RunChecks(force bool) error
	RunScript(code string, shell string, args []string, timeout int) (stdout, stderr string, exitcode int, e error)

	ScriptCheck(data shared.Check, r *resty.Client)
	DiskCheck(data shared.Check, r *resty.Client)
	CPULoadCheck(data shared.Check, r *resty.Client)
	MemCheck(data shared.Check, r *resty.Client)
	PingCheck(data shared.Check, r *resty.Client)

	OSInfo() (plat, osFullName string)
	GetDisksNATS() []trmm.Disk
	// GetDisks() []shared.Disk
	LoggedOnUser() string
	GetCPULoadAvg() int
	RecoverAgent()
	SyncInfo()
	SendSoftware()
	AgentUpdate(url, inno, version string)
	AgentUninstall()

	// new:
	Hostname() string
	AgentID() string
	// Logger() *logrus.Logger
	ShowStatus(version string)
}

// GenerateAgentID creates and returns a unique Agent ID
func GenerateAgentID() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 40)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// ShowVersionInfo prints basic debugging info
func ShowVersionInfo(ver string) {
	fmt.Println(AGENT_NAME_LONG, " Agent: ", ver)
	fmt.Println("Arch: ", runtime.GOARCH)
	fmt.Println("Go version: ", runtime.Version())
	if runtime.GOOS == "windows" {
		fmt.Println("Program Directory: ", filepath.Join(os.Getenv("ProgramFiles"), AGENT_FOLDER))
	}
}

func ShowStatus(version string) {

}
