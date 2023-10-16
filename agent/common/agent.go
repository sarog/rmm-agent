package common

import (
	"github.com/go-resty/resty/v2"
	"github.com/sarog/rmmagent/shared"
	"github.com/sarog/trmm-shared"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type IAgentLogger interface {
	setLogger(logger *logrus.Logger)
	getLogger() *logrus.Logger
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
	RunTask() // currently in BaseAgent
}

// todo
type ServiceManager interface {
	// new: Add(name string) error
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
}

type BaseAgent interface {
	// New(logger *logrus.Logger, version string) *Agent

	// Setup
	Install(i *InstallInfo, agentID string)
	AgentUpdate(url, inno, version string)
	AgentUninstall()
	UninstallCleanup()

	// Service management
	RunAgentService()
	// Deprecated todo: replace with combined Service
	RunRPCService()

	Hostname() string
	ShowStatus(version string)

	RunTask(int) error
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

type IAgent interface {
	BaseAgent
	InfoCollector

	// IAgentConfig
	// IAgentLogger
}

// test:
/*type Agent struct {
	IAgent
	// Common
	// Config *AgentConfig
	// InfoCollector
	Logger  *logrus.Logger
	RClient *resty.Client
}*/

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
