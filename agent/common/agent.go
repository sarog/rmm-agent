package common

import (
	"fmt"
	ps "github.com/elastic/go-sysinfo"
	"github.com/go-resty/resty/v2"
	jrmm "github.com/jetrmm/rmm-shared"
	"github.com/kardianos/service"
	"github.com/nats-io/nats.go"
	"github.com/oklog/ulid/v2"
	"github.com/sarog/rmmagent/agent/config"
	"github.com/sarog/rmmagent/shared"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
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
	GetInstalledSoftware() []shared.Software
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

// todo
type Messenger interface {
	SendMessage()
	ReceiveMessage()
}

type BaseAgent interface {
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

	// Windows-specific:
	// InstallUpdates(guids []string)
	// ControlService(name, action string) windows.WinSvcResp
	// EditService(name, startupType string) windows.WinSvcResp
	// GetServiceDetail(name string) shared.WindowsService
	// GetServicesNATS() []jrmm.WindowsService
	// GetServices() []shared.WindowsService
	// CreateSchedTask(st windows.SchedTask) (bool, error)
}

type IAgent interface {
	BaseAgent
	InfoCollector
	service.Interface

	// IAgentConfig
	// IAgentLogger
}

type Agent struct {
	IAgent
	*config.AgentConfig
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

func (a *Agent) SetupNatsOptions() []nats.Option {
	opts := make([]nats.Option, 0)
	opts = append(opts, nats.Name(a.AgentID))
	opts = append(opts, nats.UserInfo(a.AgentID, a.Token))
	opts = append(opts, nats.ReconnectWait(time.Second*5))
	opts = append(opts, nats.RetryOnFailedConnect(true))
	opts = append(opts, nats.MaxReconnects(-1))
	opts = append(opts, nats.ReconnectBufSize(-1))
	// opts = append(opts, nats.PingInterval(time.Duration(a.NatsPingInterval)*time.Second))
	// opts = append(opts, nats.Compression(a.NatsWSCompression))
	// opts = append(opts, nats.ProxyPath(a.NatsProxyPath))
	// opts = append(opts, nats.ReconnectJitter(500*time.Millisecond, 4*time.Second))
	// opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
	// 	a.Logger.Debugln("NATS disconnected:", err)
	// 	a.Logger.Debugf("%+v\n", nc.Statistics)
	// }))
	// opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
	// 	a.Logger.Debugln("NATS reconnected")
	// 	a.Logger.Debugf("%+v\n", nc.Statistics)
	// }))
	// opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
	// 	a.Logger.Errorln("NATS error:", err)
	// 	a.Logger.Errorf("%+v\n", sub)
	// }))
	// if a.Insecure {
	// 	insecureConf := &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	}
	// 	opts = append(opts, nats.Secure(insecureConf))
	// }
	return opts
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
