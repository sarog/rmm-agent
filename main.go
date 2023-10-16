package main

import (
	"flag"
	"fmt"
	"github.com/sarog/rmmagent/agent"
	"github.com/sarog/rmmagent/agent/windows"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

var (
	version = "0.1.0"
	log     = logrus.New()
	logFile *os.File
)

const (
	AGENT_LOG_FILE = "agent.log"

	AGENT_MODE_RPC         = "rpc"
	AGENT_MODE_SVC         = "agentsvc"
	AGENT_MODE_WINSVC      = "winagentsvc"
	AGENT_MODE_CHECKRUNNER = "checkrunner"
	AGENT_MODE_CLEANUP     = "cleanup"
	AGENT_MODE_INSTALL     = "install"
	AGENT_MODE_SHOW_PK     = "pk"
	AGENT_MODE_PUBLICIP    = "publicip"
	AGENT_MODE_RUNCHECKS   = "runchecks"
	AGENT_MODE_SOFTWARE    = "software"
	AGENT_MODE_SYNC        = "sync"
	AGENT_MODE_SYSINFO     = "sysinfo"
	AGENT_MODE_TASK        = "task"
	AGENT_MODE_TASKRUNNER  = "taskrunner"
	AGENT_MODE_UPDATE      = "update"
	AGENT_MODE_WMI         = "wmi"
)

func main() {
	hostname, _ := os.Hostname()

	// CLI
	ver := flag.Bool("version", false, "Prints agent version and exits")
	mode := flag.String("m", "", "The mode to run: "+
		"install, update, rpc, agentsvc, runchecks, checkrunner, sysinfo, software, \n\t\tsync, wmi, pk, publicip, runmigrations, taskrunner, cleanup")

	taskPK := flag.Int("p", 0, "Task PK")
	logLevel := flag.String("log", "INFO", "Log level: INFO*, WARN, ERROR, DEBUG")
	logTo := flag.String("logto", "file", "Log destination: file, stdout")

	apiUrl := flag.String("api", "", "API URL")
	clientID := flag.Int("client-id", 0, "Client ID")
	siteID := flag.Int("site-id", 0, "Site ID")
	token := flag.String("auth", "", "Agent's authorization token")

	timeout := flag.Duration("timeout", 1000, "Installer timeout in seconds")
	aDesc := flag.String("desc", hostname, "Agent's description to display on the RMM server")

	cert := flag.String("cert", "", "Path to the Certificate Authority's .pem")
	updateurl := flag.String("updateurl", "", "Source URL to retrieve the update executable")
	inno := flag.String("inno", "", "Inno setup filename")
	updatever := flag.String("updatever", "", "Update version")

	silent := flag.Bool("silent", false, "Do not popup any message boxes during installation")

	flag.Parse()

	if *ver {
		showVersionInfo(version)
		return
	}

	if len(os.Args) == 1 {
		// agent.ShowStatus(version)
		return
	}

	setupLogging(logLevel, logTo)
	defer logFile.Close()

	// was: a := *windows.New(log, version)
	a, _ := agent.New(log, version)

	switch *mode {
	case AGENT_MODE_RPC:
		a.RunRPCService()
	case AGENT_MODE_WINSVC, AGENT_MODE_SVC:
		a.RunAgentService()
	case AGENT_MODE_RUNCHECKS:
		a.RunChecks(true)
	case AGENT_MODE_CHECKRUNNER:
		a.RunChecks(false)
	case AGENT_MODE_SYSINFO, AGENT_MODE_WMI:
		a.SysInfo()
	case AGENT_MODE_SOFTWARE:
		a.SendSoftware()
	case AGENT_MODE_SYNC:
		a.SyncInfo()
	case AGENT_MODE_CLEANUP:
		a.UninstallCleanup()
	case AGENT_MODE_PUBLICIP:
		fmt.Println(a.PublicIP())
	case AGENT_MODE_TASKRUNNER, AGENT_MODE_TASK:
		if len(os.Args) < 5 || *taskPK == 0 {
			return
		}
		a.RunTask(*taskPK)
	case AGENT_MODE_SHOW_PK:
		fmt.Println(a.AgentPK)
	case AGENT_MODE_UPDATE:
		if *updateurl == "" || *inno == "" || *updatever == "" {
			updateUsage()
			return
		}
		a.AgentUpdate(*updateurl, *inno, *updatever)
	case AGENT_MODE_INSTALL:
		log.SetOutput(os.Stdout)
		if *apiUrl == "" || *clientID == 0 || *siteID == 0 || *token == "" {
			installUsage()
			return
		}
		a.Install(
			&agent.InstallInfo{
				ServerURL:   *apiUrl,
				ClientID:    *clientID,
				SiteID:      *siteID,
				Description: *aDesc,
				Token:       *token,
				Cert:        *cert,
				Timeout:     *timeout,
				Silent:      *silent,
			},
			agent.GenerateAgentID(),
		)
	default:
		a.ShowStatus(version)
	}
}

func setupLogging(level, to *string) {
	ll, err := logrus.ParseLevel(*level)
	if err != nil {
		ll = logrus.InfoLevel
	}
	log.SetLevel(ll)

	if *to == "stdout" {
		log.SetOutput(os.Stdout)
	} else {
		switch runtime.GOOS {
		case "windows":
			logFile, _ = os.OpenFile(filepath.Join(os.Getenv("ProgramFiles"), agent.AGENT_FOLDER, AGENT_LOG_FILE), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		case "freebsd":
			logFile, _ = os.OpenFile(filepath.Join("/var/log", "rmm", AGENT_LOG_FILE), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
		case "darwin":
		case "linux":
			// todo
		}
		log.SetOutput(logFile)
	}
}

func installUsage() {
	switch runtime.GOOS {
	case "windows":
		u := `Usage: %s -m install -api <https://api.example.com> -client-id X -site-id X -auth <TOKEN>`
		fmt.Printf(u, windows.AGENT_FILENAME)
	case "freebsd":
	case "darwin":
	case "linux":
		// todo
	}
}

func updateUsage() {
	switch runtime.GOOS {
	case "windows":
		u := `Usage: %s -m update -updateurl https://example.com/winagent-vX.X.X.exe -inno winagent-vX.X.X.exe -updatever 1.1.1`
		fmt.Printf(u, windows.AGENT_FILENAME)
	}
}

// showVersionInfo prints basic debugging info
func showVersionInfo(ver string) {
	fmt.Println(agent.AGENT_NAME_LONG, ver, runtime.GOARCH, runtime.Version())
	// if runtime.GOOS == "windows" {
	// 	fmt.Println("Program Directory: ", filepath.Join(os.Getenv("ProgramFiles"), agent.AGENT_FOLDER))
	// }
}
