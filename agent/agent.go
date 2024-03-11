package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	ps "github.com/jetrmm/go-sysinfo"
	"github.com/jetrmm/rmm-agent/shared"
	jrmm "github.com/jetrmm/rmm-shared"
	"github.com/kardianos/service"
	"github.com/nats-io/nats.go"
	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InfoCollector interface {
	PublicIP() string
	TotalRAM() float64
	BootTime() int64
	GetInstalledSoftware() []jrmm.Software
	OSInfo() (plat, osFullName string)
	SysInfo()
	GetStorage() []jrmm.StorageDrive
	LoggedOnUser() string
	GetCPULoadAvg() int
}

type TaskChecker interface {
	ScriptCheck(data shared.Check, r *resty.Client)
	DiskCheck(data shared.Check, r *resty.Client)
	CPULoadCheck(data shared.Check, r *resty.Client)
	MemCheck(data shared.Check, r *resty.Client)
	PingCheck(data shared.Check, r *resty.Client)
	// CheckService(data shared.X, r *resty.Client)
}

type TaskScheduler interface {
	RunTask(int) // currently in baseAgent
	CreateTask(task any) (bool, error)
	DeleteTask(task string) error
	EnableTask(task any) error
	DisableTask(task any) error
	CleanupTasks()
	ListTasks() []string
}

type PackageManager interface {
	InstallPkgMgr(mgr string)
	RemovePkgMgr(mgr string)
	InstallPackage(mgr string, name string) (string, error)
	RemovePackage(mgr string, name string) (string, error)
	UpdatePackage(mgr string, name string) (string, error)
}

// todo
type ServiceManager interface {
	// new: Add(name string) error
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
}

// Messenger is our communication interface (for RPC, JSON, etc.)
type Messenger interface {
	RpcProcessor
	Send(any)
	Receive(any)
}

type RpcProcessor interface {
	ProcessRpcMsg(conn *nats.Conn, msg *nats.Msg)
}

type baseAgent interface {
	// New(logger *logrus.Logger, version string) *Agent

	// Setup
	Install(i *InstallInfo, agentID string)
	InstallService() error
	AgentUpdate(url, inno, version string)
	AgentUninstall()
	UninstallCleanup()

	// Agent Service
	RunAgentService(conn *nats.Conn)
	RunService()

	// GetHostname() string
	ShowStatus(version string)

	RunTask(int) error
	RunChecks(force bool) error
	RunScript(code string, shell string, args []string, timeout int) (stdout, stderr string, exitcode int, e error)
	CheckIn(nc *nats.Conn, mode string)
	CreateInternalTask(name, args, repeat string, start int) (bool, error)
	CheckRunner()
	GetCheckInterval() (int, error)

	// Transmit
	SendSoftware()
	SyncInfo()

	RecoverAgent()

	GetServiceConfig() *service.Config

	RebootSystem()

	// Windows-specific:
	// InstallUpdates(guids []string)
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
	// GetServiceDetail(name string) shared.WindowsService
	// GetServicesNATS() []jrmm.WindowsService
	// GetServices() []shared.WindowsService
}

type IAgent interface {
	baseAgent
	InfoCollector
	PackageManager
	RpcProcessor // Messenger
	service.Interface

	// IAgentConfig
	// IAgentLogger
}

type Agent struct {
	IAgent
	*AgentConfig
	Logger  *logrus.Logger
	RClient *resty.Client
}

func (a *Agent) Start(s service.Service) error {
	if service.Interactive() {
		a.Logger.Info("Running in terminal.")
	} else {
		a.Logger.Info("Running under service manager.")
	}

	go a.RunService()
	return nil
}

func (a *Agent) Stop(s service.Service) error {
	a.Logger.Info("Agent service is stopping")
	return nil
}

func (a *Agent) GetHostname() string {
	sysHost, _ := ps.Host()
	return sysHost.Info().Hostname
}

// GenerateAgentID creates and returns a unique agent ULID
func GenerateAgentID() (ulid.ULID, error) {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	// fmt.Println(ulid.New(ms, entropy))

	// agentULID, err := ulid.New(ms, entropy)
	// if err != nil {
	// 	return , err
	// }

	// return agentULID, nil
	return ulid.New(ms, entropy)

	// rand.New(rand.NewSource(time.Now().UnixNano()))
	// letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	// b := make([]rune, 40)
	// for i := range b {
	// 	b[i] = letters[rand.Intn(len(letters))]
	// }
	// return string(b)
}

// CreateAgentTempDir Create the temp directory for running scripts
func (a *Agent) CreateAgentTempDir() {
	dir := filepath.Join(os.TempDir(), AGENT_TEMP_DIR)
	if !FileExists(dir) {
		// todo: 2021-12-31: verify permissions
		err := os.Mkdir(dir, 0775)
		if err != nil {
			a.Logger.Errorln(err)
		}
	}
}

// PublicIP returns the agent's public IP address
// Tries 3 times before giving up
func (a *Agent) PublicIP() string {
	a.Logger.Debugln("PublicIP start")
	client := resty.New()
	client.SetTimeout(4 * time.Second)
	// todo: 2021-12-31: allow custom URLs for IP lookups
	urls := []string{"https://icanhazip.com", "https://ifconfig.co/ip"}
	ip := "error"

	for _, url := range urls {
		r, err := client.R().Get(url)
		if err != nil {
			a.Logger.Debugln("PublicIP error", err)
			continue
		}
		ip = StripAll(r.String())
		if !IsValidIP(ip) {
			a.Logger.Debugln("PublicIP not valid", ip)
			continue
		}
		v4 := net.ParseIP(ip)
		if v4.To4() == nil {
			r1, err := client.R().Get("https://ifconfig.me/ip")
			if err != nil {
				return ip
			}
			ipv4 := StripAll(r1.String())
			if !IsValidIP(ipv4) {
				continue
			}
			a.Logger.Debugln("Forcing IPv4:", ipv4)
			return ipv4
		}
		a.Logger.Debugln("PublicIP return: ", ip)
		break
	}
	return ip
}

// TotalRAM returns total RAM in GB
func (a *Agent) TotalRAM() float64 {
	host, err := ps.Host()
	if err != nil {
		return 8.0
	}
	mem, err := host.Memory()
	if err != nil {
		return 8.0
	}
	return math.Ceil(float64(mem.Total) / 1073741824.0)
}

// BootTime returns system boot time as a Unix timestamp
func (a *Agent) BootTime() int64 {
	host, err := ps.Host()
	if err != nil {
		return 1000
	}
	info := host.Info()
	return info.BootTime.Unix()
}

// OSInfo returns formatted OS names
func (a *Agent) OSInfo() (plat, osFullName string) {
	host, _ := ps.Host()
	info := host.Info()
	osInfo := info.OS

	var arch string
	switch info.Architecture {
	case "x86_64":
		arch = "64 bit"
	case "x86":
		arch = "32 bit"
	}

	plat = osInfo.Platform
	osFullName = fmt.Sprintf("%s, %s (build %s)", osInfo.Name, arch, osInfo.Build)
	return
}

func TestTCP(addr string) error {
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// IsValidIP checks for a valid IPv4 or IPv6 address
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// StripAll strips all whitespace and newline chars
func StripAll(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\n")
	s = strings.Trim(s, "\r")
	return s
}
