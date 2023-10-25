package windows

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/gonutz/w32/v2"
	jrmm "github.com/jetrmm/rmm-shared"
	"github.com/kardianos/service"
	"github.com/sarog/rmmagent/agent/common"
	"github.com/sarog/rmmagent/agent/config"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	ps "github.com/elastic/go-sysinfo"
	wapf "github.com/jetrmm/go-win64api"
	rmm "github.com/sarog/rmmagent/shared"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	getDriveType = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetDriveTypeW")
)

const (
	AGENT_FILENAME     = "rmmagent.exe"
	INNO_SETUP_DIR     = "rmmagent"
	INNO_SETUP_LOGFILE = "rmmagent.txt"
	AGENT_MODE_COMMAND = "command"

	API_URL_SOFTWARE = "/api/v3/software/"
)

type windowsAgent struct {
	common.Agent
	ProgramDir  string
	AgentExe    string
	SystemDrive string // needed?
}

func NewAgent(logger *logrus.Logger, version string) common.IAgent {
	host, _ := ps.Host()
	info := host.Info()
	pd := filepath.Join(os.Getenv("ProgramFiles"), AGENT_FOLDER)
	exe := filepath.Join(pd, AGENT_FILENAME)
	sd := os.Getenv("SystemDrive")

	var (
		baseurl string
		agentid string
		apiurl  string
		token   string
		agentpk string
		pk      int
		cert    string
	)

	// todo: 2021-12-31: migrate to DPAPI
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, REG_RMM_PATH, registry.ALL_ACCESS)
	if err == nil {
		baseurl, _, err = key.GetStringValue(REG_RMM_BASEURL)
		if err != nil {
			logger.Fatalln("Unable to get BaseURL:", err)
		}

		agentid, _, err = key.GetStringValue(REG_RMM_AGENTID)
		if err != nil {
			logger.Fatalln("Unable to get AgentID:", err)
		}

		apiurl, _, err = key.GetStringValue(REG_RMM_APIURL)
		if err != nil {
			logger.Fatalln("Unable to get ApiURL:", err)
		}

		token, _, err = key.GetStringValue(REG_RMM_TOKEN)
		if err != nil {
			logger.Fatalln("Unable to get Token:", err)
		}

		agentpk, _, err = key.GetStringValue(REG_RMM_AGENTPK)
		if err != nil {
			logger.Fatalln("Unable to get AgentPK:", err)
		}

		pk, _ = strconv.Atoi(agentpk)

		cert, _, _ = key.GetStringValue(REG_RMM_CERT)
	}

	headers := make(map[string]string)
	if len(token) > 0 {
		headers["Content-Type"] = "application/json"
		headers["Authorization"] = fmt.Sprintf("Token %s", token)
	}

	restyC := resty.New()
	restyC.SetBaseURL(baseurl)
	restyC.SetCloseConnection(true)
	restyC.SetHeaders(headers)
	restyC.SetTimeout(15 * time.Second)
	restyC.SetDebug(logger.IsLevelEnabled(logrus.DebugLevel))
	if len(cert) > 0 {
		restyC.SetRootCertificate(cert)
	}

	return &windowsAgent{
		Agent: common.Agent{
			AgentConfig: &config.AgentConfig{
				AgentID:  agentid,
				BaseURL:  baseurl,
				ApiURL:   apiurl,
				ApiPort:  common.NATS_DEFAULT_PORT,
				Token:    token,
				AgentPK:  pk,
				Cert:     cert,
				Arch:     info.Architecture,
				Hostname: info.Hostname,
				Version:  version,
				Debug:    logger.IsLevelEnabled(logrus.DebugLevel),
				Headers:  headers,
			},
			Logger:  logger,
			RClient: restyC,
		},
		ProgramDir:  pd,
		AgentExe:    exe,
		SystemDrive: sd,
	}
}

// New Initializes a new windowsAgent with logger
func (a *windowsAgent) New(logger *logrus.Logger, version string) *windowsAgent {
	// return &NewAgent(logger, version)

	host, _ := ps.Host()
	info := host.Info()
	pd := filepath.Join(os.Getenv("ProgramFiles"), AGENT_FOLDER)
	exe := filepath.Join(pd, AGENT_FILENAME)
	sd := os.Getenv("SystemDrive")

	var (
		baseurl string
		agentid string
		apiurl  string
		token   string
		agentpk string
		pk      int
		cert    string
	)

	// todo: 2021-12-31: migrate to DPAPI
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, REG_RMM_PATH, registry.ALL_ACCESS)
	if err == nil {
		baseurl, _, err = key.GetStringValue(REG_RMM_BASEURL)
		if err != nil {
			logger.Fatalln("Unable to get BaseURL:", err)
		}

		agentid, _, err = key.GetStringValue(REG_RMM_AGENTID)
		if err != nil {
			logger.Fatalln("Unable to get AgentID:", err)
		}

		apiurl, _, err = key.GetStringValue(REG_RMM_APIURL)
		if err != nil {
			logger.Fatalln("Unable to get ApiURL:", err)
		}

		token, _, err = key.GetStringValue(REG_RMM_TOKEN)
		if err != nil {
			logger.Fatalln("Unable to get Token:", err)
		}

		agentpk, _, err = key.GetStringValue(REG_RMM_AGENTPK)
		if err != nil {
			logger.Fatalln("Unable to get AgentPK:", err)
		}

		pk, _ = strconv.Atoi(agentpk)

		cert, _, _ = key.GetStringValue(REG_RMM_CERT)
	}

	headers := make(map[string]string)
	if len(token) > 0 {
		headers["Content-Type"] = "application/json"
		headers["Authorization"] = fmt.Sprintf("Token %s", token)
	}

	restyC := resty.New()
	restyC.SetBaseURL(baseurl)
	restyC.SetCloseConnection(true)
	restyC.SetHeaders(headers)
	restyC.SetTimeout(15 * time.Second)
	restyC.SetDebug(logger.IsLevelEnabled(logrus.DebugLevel))
	if len(cert) > 0 {
		restyC.SetRootCertificate(cert)
	}

	return &windowsAgent{
		Agent: common.Agent{
			AgentConfig: &config.AgentConfig{
				AgentID:  agentid,
				BaseURL:  baseurl,
				ApiURL:   apiurl,
				ApiPort:  common.NATS_DEFAULT_PORT,
				Token:    token,
				AgentPK:  pk,
				Cert:     cert,
				Arch:     info.Architecture,
				Hostname: info.Hostname,
				Version:  version,
				Debug:    logger.IsLevelEnabled(logrus.DebugLevel),
				Headers:  headers,
			},
			Logger:  logger,
			RClient: restyC,
		},
		ProgramDir:  pd,
		AgentExe:    exe,
		SystemDrive: sd,
	}
}

// OSInfo returns formatted OS names
func (a *windowsAgent) OSInfo() (plat, osFullName string) {
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

// GetStorage returns a list of fixed disks
func (a *windowsAgent) GetStorage() []jrmm.Disk {
	ret := make([]jrmm.Disk, 0)
	partitions, err := disk.Partitions(false)
	if err != nil {
		a.Logger.Debugln(err)
		return ret
	}

	for _, p := range partitions {
		typepath, _ := windows.UTF16PtrFromString(p.Device)
		typeval, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(typepath)))
		// https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getdrivetypea
		if typeval != 3 {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			a.Logger.Debugln(err)
			continue
		}

		d := jrmm.Disk{
			Device:  p.Device,
			Fstype:  p.Fstype,
			Total:   string(usage.Total),
			Used:    string(usage.Used),
			Free:    string(usage.Free),
			Percent: int(usage.UsedPercent),
		}
		ret = append(ret, d)
	}
	return ret
}

// GetDisks returns a list of fixed disks
// Deprecated, use GetStorage()
func (a *windowsAgent) GetDisks() []rmm.Storage {
	ret := make([]rmm.Storage, 0)
	partitions, err := disk.Partitions(false)
	if err != nil {
		a.Logger.Debugln(err)
		return ret
	}

	for _, p := range partitions {
		typepath, _ := windows.UTF16PtrFromString(p.Device)
		typeval, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(typepath)))
		// https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getdrivetypea
		if typeval != 3 {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			a.Logger.Debugln(err)
			continue
		}

		d := rmm.Storage{
			Device:  p.Device,
			Fstype:  p.Fstype,
			Total:   usage.Total,
			Used:    usage.Used,
			Free:    usage.Free,
			Percent: usage.UsedPercent,
		}
		ret = append(ret, d)
	}
	return ret
}

func CMDShell(shell string, cmdArgs []string, command string, timeout int, detached bool) (output [2]string, e error) {
	var (
		outb     bytes.Buffer
		errb     bytes.Buffer
		cmd      *exec.Cmd
		timedOut bool = false
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// sysProcAttr := &windows.SysProcAttr{}

	if len(cmdArgs) > 0 && command == "" {
		switch shell {
		case "cmd":
			cmdArgs = append([]string{"/C"}, cmdArgs...)
			cmd = exec.Command("cmd.exe", cmdArgs...)
		case "powershell":
			cmdArgs = append([]string{"-NonInteractive", "-NoProfile"}, cmdArgs...)
			cmd = exec.Command("powershell.exe", cmdArgs...)
		}
	} else {
		switch shell {
		case "cmd":
			cmd = exec.Command("cmd.exe")
			cmd.SysProcAttr = &windows.SysProcAttr{
				CmdLine: fmt.Sprintf("cmd.exe /C %s", command),
			}
		case "powershell":
			cmd = exec.Command("Powershell", "-NonInteractive", "-NoProfile", command)
		}
	}

	// https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
	if detached {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
		}
	}

	// https://blog.davidvassallo.me/2022/06/17/golang-in-windows-execute-command-as-another-user/
	/*if runAsUser {
		token, err := wintoken.GetInteractiveToken(wintoken.TokenImpersonation)
		if err != nil {
			return [2]string{"", CleanString(err.Error())}, err
		}
		defer token.Close()
		sysProcAttr.Token = syscall.Token(token.Token())
		sysProcAttr.HideWindow = true
	}*/

	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Start()

	pid := int32(cmd.Process.Pid)

	go func(p int32) {
		<-ctx.Done()
		_ = common.KillProc(p)
		timedOut = true
	}(pid)

	err = cmd.Wait()

	if timedOut {
		return [2]string{outb.String(), errb.String()}, ctx.Err()
	}

	if err != nil {
		return [2]string{outb.String(), errb.String()}, err
	}

	return [2]string{outb.String(), errb.String()}, nil
}

// CMD runs a command with shell=False
func CMD(exe string, args []string, timeout int, detached bool) (output [2]string, e error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var outb, errb bytes.Buffer
	cmd := exec.CommandContext(ctx, exe, args...)
	if detached {
		cmd.SysProcAttr = &windows.SysProcAttr{
			CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
		}
	}
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		return [2]string{"", ""}, fmt.Errorf("%s: %s", err, errb.String())
	}

	if ctx.Err() == context.DeadlineExceeded {
		return [2]string{"", ""}, ctx.Err()
	}

	return [2]string{outb.String(), errb.String()}, nil
}

// LoggedOnUser returns the first logged on user it finds
// todo: return more than 1 user (e.g. RDS servers)
func (a *windowsAgent) LoggedOnUser() string {
	// 2022-01-02: Works in PowerShell 5.x and Core 7.x
	cmd := "((Get-CimInstance -ClassName Win32_ComputerSystem).Username).Split('\\')[1]"
	user, _, _, err := a.RunScript(cmd, "powershell", []string{}, 20)
	if err != nil {
		a.Logger.Debugln(err)
	}
	if err == nil {
		return user
	}

	// Attempt #2: Go fallback
	users, err := wapf.ListLoggedInUsers()
	if err != nil {
		a.Logger.Debugln("LoggedOnUser error", err)
		return "None"
	}

	if len(users) == 0 {
		return "None"
	}

	for _, u := range users {
		// Strip the 'Domain\' (or 'ComputerName\') prefix
		return strings.Split(u.FullUser(), `\`)[1]
	}
	return "None"
}

// GetCPULoadAvg Retrieve CPU load average
func (a *windowsAgent) GetCPULoadAvg() int {
	fallback := false

	// 2022-01-02: Works in PowerShell 5.x and Core 7.x
	// todo? | Measure-Object -Property LoadPercentage -Average | Select Average
	cmd := "(Get-CimInstance -ClassName Win32_Processor).LoadPercentage"
	load, _, _, err := a.RunScript(cmd, "powershell", []string{}, 20)

	if err != nil {
		a.Logger.Debugln(err)
		fallback = true
	}

	i, _ := strconv.Atoi(load)

	if fallback {
		percent, err := cpu.Percent(10*time.Second, false)
		if err != nil {
			a.Logger.Debugln("Go CPU Check:", err)
			return 0
		}
		return int(math.Round(percent[0]))
	}
	return i
}

// RecoverAgent Recover the Agent
func (a *windowsAgent) RecoverAgent() {
	a.Logger.Debugln("Attempting ", common.AGENT_NAME_LONG, " recovery on", a.Hostname)

	// a.Logger.Infoln("Attempting agent service recovery")
	_, _ = CMD("net", []string{"stop", SERVICE_NAME_RPC}, 90, false)
	time.Sleep(2 * time.Second)
	_, _ = CMD("net", []string{"start", SERVICE_NAME_RPC}, 90, false)

	// todo? a.Restart()

	// defer CMD(a.Nssm, []string{"start", SERVICE_NAME_AGENT}, 60, false)
	// _, _ = CMD(a.Nssm, []string{"stop", SERVICE_NAME_AGENT}, 120, false)
	_, _ = CMD("ipconfig", []string{"/flushdns"}, 15, false)
	a.Logger.Debugln(common.AGENT_NAME_LONG, " recovery completed on", a.Hostname)
}

// RecoverCMD runs a shell recovery command
func (a *windowsAgent) RecoverCMD(command string) {
	a.Logger.Infoln("Attempting shell recovery with command:", command)
	// To prevent killing ourselves, prefix the command with 'cmd /C'
	// so the parent process is now cmd.exe and not rmmagent.exe
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
		CmdLine:       fmt.Sprintf("cmd.exe /C %s", command), // properly escape in case double quotes are in the command
	}
	cmd.Start()
}

func (a *windowsAgent) SyncInfo() {
	a.SysInfo()
	time.Sleep(1 * time.Second)
	a.SendSoftware()
}

// SendSoftware Send list of installed software
func (a *windowsAgent) SendSoftware() {
	sw := a.GetInstalledSoftware()
	a.Logger.Debugln(sw)

	payload := map[string]interface{}{
		"agent_id": a.AgentID,
		"software": sw,
	}

	_, err := a.RClient.R().SetBody(payload).Post(API_URL_SOFTWARE)
	if err != nil {
		a.Logger.Debugln(err)
	}
}

func (a *windowsAgent) UninstallCleanup() {
	registry.DeleteKey(registry.LOCAL_MACHINE, REG_RMM_PATH)
	a.CleanupAgentUpdates()
	CleanupSchedTasks()
}

// ShowStatus prints the Windows service status
// If called from an interactive desktop, pops up a message box
// Otherwise prints to the console
func (a *windowsAgent) ShowStatus(version string) {
	statusMap := make(map[string]string)
	svcs := []string{SERVICE_NAME_AGENT}
	// was: svcs := []string{SERVICE_NAME_AGENT, SERVICE_NAME_RPC}

	for _, svc := range svcs {
		status, err := GetServiceStatus(svc)
		if err != nil {
			statusMap[svc] = "Not Installed"
			continue
		}
		statusMap[svc] = status
	}

	// todo: if service.Interactive()

	window := w32.GetForegroundWindow()
	if window != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(window)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindow(window, w32.SW_HIDE)
		}
		var handle w32.HWND
		msg := fmt.Sprintf("Agent Service: %s", statusMap[SERVICE_NAME_AGENT])

		// was: msg := fmt.Sprintf("Agent: %s\n\nRPC Service: %s", statusMap[SERVICE_NAME_AGENT], statusMap[SERVICE_NAME_RPC])

		w32.MessageBox(handle, msg, fmt.Sprintf("%s v%s", common.AGENT_NAME_LONG, version), w32.MB_OK|w32.MB_ICONINFORMATION)
	} else {
		fmt.Println("Agent Version", version)
		fmt.Println("Agent Service:", statusMap[SERVICE_NAME_AGENT])
		// fmt.Println("RPC Service:", statusMap[SERVICE_NAME_RPC])
	}
}

func (a *windowsAgent) installerMsg(msg, alert string, silent bool) {
	window := w32.GetForegroundWindow()
	if !silent && window != 0 {
		var (
			handle w32.HWND
			flags  uint
		)

		switch alert {
		case "info":
			flags = w32.MB_OK | w32.MB_ICONINFORMATION
		case "error":
			flags = w32.MB_OK | w32.MB_ICONERROR
		default:
			flags = w32.MB_OK | w32.MB_ICONINFORMATION
		}

		w32.MessageBox(handle, msg, common.AGENT_NAME_LONG, flags)
	} else {
		fmt.Println(msg)
	}

	if alert == "error" {
		a.Logger.Fatalln(msg)
	}
}

func (a *windowsAgent) AgentUpdate(url, inno, version string) {
	time.Sleep(time.Duration(randRange(1, 15)) * time.Second)

	a.CleanupAgentUpdates()

	updater := filepath.Join(a.ProgramDir, inno)
	a.Logger.Infof("Agent updating from %s to %s", a.Version, version)
	a.Logger.Infoln("Downloading agent update from", url)

	rClient := resty.New()
	rClient.SetCloseConnection(true)
	rClient.SetTimeout(15 * time.Minute)
	rClient.SetDebug(a.Debug)
	r, err := rClient.R().SetOutput(updater).Get(url)
	if err != nil {
		a.Logger.Errorln(err)
		CMD("net", []string{"start", SERVICE_NAME_RPC}, 10, false)
		return
	}
	if r.IsError() {
		a.Logger.Errorln("Download failed with status code", r.StatusCode())
		CMD("net", []string{"start", SERVICE_NAME_RPC}, 10, false)
		return
	}

	dir, err := os.MkdirTemp("", INNO_SETUP_DIR)
	if err != nil {
		a.Logger.Errorln("AgentUpdate unable to create temporary directory:", err)
		CMD("net", []string{"start", SERVICE_NAME_RPC}, 10, false)
		return
	}

	innoLogFile := filepath.Join(dir, INNO_SETUP_LOGFILE)

	args := []string{"/C", updater, "/VERYSILENT", fmt.Sprintf("/LOG=%s", innoLogFile)}
	cmd := exec.Command("cmd.exe", args...)
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Start()
	time.Sleep(1 * time.Second)
}

func (a *windowsAgent) GetUninstallExe() string {
	cderr := os.Chdir(a.ProgramDir)
	if cderr == nil {
		files, err := filepath.Glob("unins*.exe")
		if err == nil {
			for _, f := range files {
				if strings.Contains(f, "001") {
					return f
				}
			}
		}
	}
	return "unins000.exe"
}

func (a *windowsAgent) AgentUninstall() {
	agentUninst := filepath.Join(a.ProgramDir, a.GetUninstallExe())
	args := []string{"/C", agentUninst, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/FORCECLOSEAPPLICATIONS"}
	cmd := exec.Command("cmd.exe", args...)
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Start()
}

func (a *windowsAgent) CleanupAgentUpdates() {
	cderr := os.Chdir(a.ProgramDir)
	if cderr != nil {
		a.Logger.Errorln(cderr)
		return
	}

	files, err := filepath.Glob("winagent-v*.exe")
	if err == nil {
		for _, f := range files {
			os.Remove(f)
		}
	}

	cderr = os.Chdir(os.Getenv("TMP"))
	if cderr != nil {
		a.Logger.Errorln(cderr)
		return
	}
	folders, err := filepath.Glob(RMM_SEARCH_PREFIX)
	if err == nil {
		for _, f := range folders {
			os.RemoveAll(f)
		}
	}
}

// CheckForRecovery Check for agent recovery signal
func (a *windowsAgent) CheckForRecovery() {
	url := fmt.Sprintf("/api/v3/%s/recovery/", a.AgentID)
	r, err := a.RClient.R().SetResult(&rmm.RecoveryAction{}).Get(url)

	if err != nil {
		a.Logger.Debugln("Recovery:", err)
		return
	}
	if r.IsError() {
		a.Logger.Debugln("Recovery status code:", r.StatusCode())
		return
	}

	mode := r.Result().(*rmm.RecoveryAction).Mode
	command := r.Result().(*rmm.RecoveryAction).ShellCMD

	switch mode {
	case AGENT_MODE_RPC:
		a.RecoverAgent()
	case AGENT_MODE_COMMAND:
		a.RecoverCMD(command)
	default:
		return
	}
}

func (a *windowsAgent) GetServiceConfig() *service.Config {
	return &service.Config{
		Name:        SERVICE_NAME_AGENT,
		DisplayName: SERVICE_DISP_AGENT,
		Description: SERVICE_DESC_AGENT,
		Executable:  AGENT_FILENAME,
		Arguments:   []string{"-m", AGENT_MODE_SVC},
		Option: service.KeyValue{
			"StartType":              "automatic",
			"OnFailure":              "restart",
			"OnFailureDelayDuration": SERVICE_RESTART_DELAY,
		},
	}
}

/*func (a *windowsAgent) getToken(pid int) (syscall.Token, error) {
	var err error
	var token syscall.Token

	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		a.Logger.Warn("Token Process", "err", err)
	}
	defer syscall.CloseHandle(handle)

	// Find process token via win32
	err = syscall.OpenProcessToken(handle, syscall.TOKEN_ALL_ACCESS, &token)

	if err != nil {
		a.Logger.Warn("Open Token Process", "err", err)
	}
	return token, err
}

func (a *windowsAgent) runAsUser(cmdPath string) (string, string) {
	token, err := a.getToken()
	if err != nil {
		a.Logger.Warn("Get Token", "err", err)
	}

	defer token.Close()

	cmd := exec.Command(cmdPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Token: token,
	}

	err = cmd.Run()

	return stdout.String(), stderr.String()
}*/
