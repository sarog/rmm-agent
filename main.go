package main

import (
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"github.com/sarog/rmmagent/agent"
	"github.com/sarog/rmmagent/agent/common"
	"github.com/sarog/rmmagent/agent/windows"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
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
)

func main() {
	hostname, _ := os.Hostname()

	help := "Missing sub-command" // todo

	// CLI
	ver := flag.Bool("version", false, "Prints agent version and exits")

	// Install
	installSet := flag.NewFlagSet("install", flag.ContinueOnError)
	silent := installSet.Bool("silent", false, "Do not popup any message boxes during installation")
	apiUrl := installSet.String("api", "", "API URL")
	clientID := installSet.Int("client-id", 0, "Client ID")
	siteID := installSet.Int("site-id", 0, "Site ID")
	token := installSet.String("auth", "", "Agent's authorization token")
	timeout := installSet.Duration("timeout", 1000, "Installer timeout in seconds")
	aDesc := installSet.String("desc", hostname, "Agent's description to display on the RMM server")
	cert := installSet.String("cert", "", "Path to the Root Certificate Authority's .pem")

	// Update
	updateSet := flag.NewFlagSet("update", flag.ContinueOnError)
	updateurl := updateSet.String("updateurl", "", "Source URL to retrieve the update executable")
	inno := updateSet.String("inno", "", "Setup filename")
	updatever := updateSet.String("updatever", "", "Update version")

	modeSet := flag.NewFlagSet("mode", flag.ContinueOnError)
	mode := modeSet.String("m", "", "The mode to run: "+
		"install, update, agentsvc, runchecks, checkrunner, sysinfo, software, \n\t\tsync, pk, publicip, taskrunner, cleanup")

	taskPK := flag.Int("p", 0, "Task PK")

	// Logging
	logLevel := flag.String("log", "INFO", "Log level: INFO*, WARN, ERROR, DEBUG")
	logTo := flag.String("logto", "file", "Log destination: file, stdout")

	// Agent Service management
	svcFlag := flag.String("service", "", "Control the system service.")

	// flag.Parse()

	// info, ok := debug.ReadBuildInfo()
	// if !ok {
	// 	fmt.Fprintln(os.Stderr, "build information not found")
	// 	return
	// }

	// if *ver {
	// 	printVersionInfo(info)
	// 	return
	// }

	if *ver {
		showVersionInfo(version)
		return
	}

	setupLogging(logLevel, logTo)
	defer logFile.Close()

	var a = agent.GetAgent(log, version).(common.IAgent)
	s, _ := service.New(a, a.GetServiceConfig())

	if len(os.Args) == 1 {
		a.ShowStatus(version)
		fmt.Fprintln(os.Stderr, "didn't receive any arguments")
		os.Exit(0)
		return
	}

	switch os.Args[1] {
	case "install":
		if err := installSet.Parse(os.Args[2:]); err == nil {
			// fmt.Println("install", "silent=", *silent, "api=", *apiUrl, "client=", *clientID, "site=", *siteID, "token=", *token, "cert", "timeout", "desc")
			installSet.PrintDefaults()
		}

	case "update":
		if err := updateSet.Parse(os.Args[2:]); err == nil {
			fmt.Println("Update the agent.")
			updateSet.PrintDefaults()
		}

	case "mode":
		if err := modeSet.Parse(os.Args[2:]); err == nil {
			fmt.Println("mode", *silent)
			modeSet.PrintDefaults()
		}

	case "service":
		fmt.Fprintln(os.Stderr, "case => service")
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}

	default:
		fmt.Fprintln(os.Stderr, "case => default")
		fmt.Fprintln(os.Stderr, help)
		os.Exit(0)
	}

	if *svcFlag != "" {
		if len(*svcFlag) != 0 {
			err := service.Control(s, *svcFlag)
			if err != nil {
				log.Printf("Valid actions: %q\n", service.ControlAction)
				log.Fatal(err)
			}
			return
		}
	}

	switch *mode {
	// case AGENT_MODE_RPC:
	// 	a.RunService()
	case AGENT_MODE_RPC, AGENT_MODE_SVC:

		s.Run()
		// a.RunService()
		// a.RunAgentService()
	case AGENT_MODE_RUNCHECKS:
		a.RunChecks(true)
	case AGENT_MODE_CHECKRUNNER:
		a.RunChecks(false)
	case AGENT_MODE_SYSINFO:
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
	// todo:
	// case AGENT_MODE_SHOW_PK:
	// 	fmt.Println(a.AgentPK)

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
			&common.InstallInfo{
				ServerURL:   *apiUrl,
				ClientID:    *clientID,
				SiteID:      *siteID,
				Description: *aDesc,
				Token:       *token,
				RootCert:    *cert,
				Timeout:     *timeout,
				Silent:      *silent,
			},
			common.GenerateAgentID(),
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
			logFile, _ = os.OpenFile(filepath.Join(os.Getenv("ProgramFiles"), windows.AGENT_FOLDER, AGENT_LOG_FILE), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		case "freebsd":
			logFile, _ = os.OpenFile(filepath.Join("/var/log", "rmm", AGENT_LOG_FILE), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
		case "darwin":
		case "linux":

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
	fmt.Println(common.AGENT_NAME_LONG, ver, runtime.GOARCH, runtime.Version())
	// if runtime.GOOS == "windows" {
	// 	fmt.Println("Program Directory: ", filepath.Join(os.Getenv("ProgramFiles"), agent.AGENT_FOLDER))
	// }
}
