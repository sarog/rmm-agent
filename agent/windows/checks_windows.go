package windows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sarog/rmmagent/agent/common"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	ps "github.com/jetrmm/go-sysinfo"
	rmm "github.com/sarog/rmmagent/shared"
	"github.com/shirou/gopsutil/v3/disk"
)

const (
	// todo: 2022-01-02: consolidate these elsewhere
	API_URL_CHECKRUNNER = "/api/v3/checkrunner/"

	// Check Types
	CHECK_TYPE_DISKSPACE = "diskspace"
	CHECK_TYPE_CPULOAD   = "cpuload"
	CHECK_TYPE_MEMORY    = "memory"
	CHECK_TYPE_PING      = "ping"
	CHECK_TYPE_SCRIPT    = "script"
	CHECK_TYPE_WINSVC    = "winsvc"
	CHECK_TYPE_EVENTLOG  = "eventlog"

	// Agent Modes
	// AGENT_MODE_CHECKRUNNER = "checkrunner"
)

func (a *windowsAgent) CheckRunner() {
	a.Logger.Infoln("CheckRunner service started.")
	sleepDelay := randRange(14, 22)
	a.Logger.Debugf("Sleeping for %v seconds", sleepDelay)
	time.Sleep(time.Duration(sleepDelay) * time.Second)
	for {
		interval, err := a.GetCheckInterval()
		if err == nil && !a.ChecksRunning() {
			_, err = CMD(a.AgentExe, []string{"-m", "checkrunner"}, 600, false)
			if err != nil {
				a.Logger.Errorln("CheckRunner RunChecks", err)
			}
		}
		a.Logger.Debugf("CheckRunner sleeping for %d seconds", interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (a *windowsAgent) GetCheckInterval() (int, error) {
	r, err := a.RClient.R().SetResult(&rmm.CheckInfo{}).Get(fmt.Sprintf("/api/v3/%s/checkinterval/", a.AgentID))
	if err != nil {
		a.Logger.Debugln(err)
		return 120, err
	}
	if r.IsError() {
		a.Logger.Debugln("CheckInterval response code:", r.StatusCode())
		return 120, fmt.Errorf("checkinterval response code: %v", r.StatusCode())
	}
	interval := r.Result().(*rmm.CheckInfo).Interval
	return interval, nil
}

func (a *windowsAgent) RunChecks(force bool) error {
	data := rmm.AllChecks{}
	var url string
	if force {
		url = fmt.Sprintf("/api/v3/%s/runchecks/", a.AgentID)
	} else {
		url = fmt.Sprintf("/api/v3/%s/checkrunner/", a.AgentID)
	}

	r, err := a.RClient.R().Get(url)
	if err != nil {
		a.Logger.Debugln(err)
		return err
	}

	if r.IsError() {
		a.Logger.Debugln("CheckRunner response code:", r.StatusCode())
		return nil
	}

	if err := json.Unmarshal(r.Body(), &data); err != nil {
		a.Logger.Debugln(err)
		return err
	}

	var wg sync.WaitGroup
	eventLogChecks := make([]rmm.Check, 0)
	winServiceChecks := make([]rmm.Check, 0)

	for _, check := range data.Checks {
		switch check.CheckType {
		case CHECK_TYPE_DISKSPACE:
			wg.Add(1)
			go func(c rmm.Check, wg *sync.WaitGroup, r *resty.Client) {
				defer wg.Done()
				time.Sleep(time.Duration(randRange(300, 950)) * time.Millisecond)
				a.DiskCheck(c, r)
			}(check, &wg, a.RClient)
		case CHECK_TYPE_CPULOAD:
			wg.Add(1)
			go func(c rmm.Check, wg *sync.WaitGroup, r *resty.Client) {
				defer wg.Done()
				a.CPULoadCheck(c, r)
			}(check, &wg, a.RClient)
		case CHECK_TYPE_MEMORY:
			wg.Add(1)
			go func(c rmm.Check, wg *sync.WaitGroup, r *resty.Client) {
				defer wg.Done()
				time.Sleep(time.Duration(randRange(300, 950)) * time.Millisecond)
				a.MemCheck(c, r)
			}(check, &wg, a.RClient)
		case CHECK_TYPE_PING:
			wg.Add(1)
			go func(c rmm.Check, wg *sync.WaitGroup, r *resty.Client) {
				defer wg.Done()
				time.Sleep(time.Duration(randRange(300, 950)) * time.Millisecond)
				a.PingCheck(c, r)
			}(check, &wg, a.RClient)
		case CHECK_TYPE_SCRIPT:
			wg.Add(1)
			go func(c rmm.Check, wg *sync.WaitGroup, r *resty.Client) {
				defer wg.Done()
				time.Sleep(time.Duration(randRange(300, 950)) * time.Millisecond)
				a.ScriptCheck(c, r)
			}(check, &wg, a.RClient)
		case CHECK_TYPE_WINSVC:
			winServiceChecks = append(winServiceChecks, check)
		case CHECK_TYPE_EVENTLOG:
			eventLogChecks = append(eventLogChecks, check)
		default:
			continue
		}
	}

	if len(winServiceChecks) > 0 {
		wg.Add(len(winServiceChecks))
		go func(wg *sync.WaitGroup, r *resty.Client) {
			for _, winSvcCheck := range winServiceChecks {
				defer wg.Done()
				a.CheckService(winSvcCheck, r)
			}
		}(&wg, a.RClient)
	}

	if len(eventLogChecks) > 0 {
		wg.Add(len(eventLogChecks))
		go func(wg *sync.WaitGroup, r *resty.Client) {
			for _, evtCheck := range eventLogChecks {
				defer wg.Done()
				a.EventLogCheck(evtCheck, r)
			}
		}(&wg, a.RClient)
	}
	wg.Wait()
	return nil
}

func (a *windowsAgent) RunScript(code string, interpreter string, args []string, timeout int) (stdout, stderr string, exitcode int, e error) {
	content := []byte(code)

	dir := filepath.Join(os.TempDir(), common.AGENT_TEMP_DIR)
	if !common.FileExists(dir) {
		a.CreateAgentTempDir()
	}

	const defaultExitCode = 1

	var (
		outb    bytes.Buffer
		errb    bytes.Buffer
		exe     string
		ext     string
		cmdArgs []string
	)

	switch interpreter {
	case "powershell":
		ext = "*.ps1"
	case "cmd":
		ext = "*.bat" // todo: .cmd?
	}

	tmpfn, err := os.CreateTemp(dir, ext)
	if err != nil {
		a.Logger.Errorln(err)
		return "", err.Error(), 85, err
	}
	defer os.Remove(tmpfn.Name())

	if _, err := tmpfn.Write(content); err != nil {
		a.Logger.Errorln(err)
		return "", err.Error(), 85, err
	}
	if err := tmpfn.Close(); err != nil {
		a.Logger.Errorln(err)
		return "", err.Error(), 85, err
	}

	switch interpreter {
	case "powershell":
		exe = "Powershell"
		// todo: 2021-12-31: allow ExecutionPolicy to be chosen by the sysadmin
		cmdArgs = []string{"-NonInteractive", "-NoProfile", "-ExecutionPolicy", "Bypass", tmpfn.Name()}
	case "cmd":
		exe = tmpfn.Name()
	}

	if len(args) > 0 {
		cmdArgs = append(cmdArgs, args...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	var timedOut bool = false
	// var token *wintoken.Token // for RunAsUser
	// var envBlock *uint16      // for Environmental Variables
	cmd := exec.Command(exe, cmdArgs...)

	// https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-createprocessasusera
	// https://github.com/gorpher/gowin32/blob/master/wrappers/userenv.go
	// if runAsUser {
	// 	token, err = wintoken.GetInteractiveToken(wintoken.TokenImpersonation)
	// 	if err == nil {
	// 		defer token.Close()
	// 		cmd.SysProcAttr = &syscall.SysProcAttr{Token: syscall.Token(token.Token()), HideWindow: true}
	//
	// 		/*if usingEnvVars {
	// 			envBlock, err = CreateEnvironmentBlock(syscall.Token(token.Token()))
	// 			if err == nil {
	// 				defer DestroyEnvironmentBlock(envBlock)
	// 				userEnv := EnvironmentBlockToSlice(envBlock)
	// 				cmd.Env = userEnv
	// 			} else {
	// 				cmd.Env = os.Environ()
	// 			}
	// 		}*/
	// 	}
	// } /*else if usingEnvVars {
	// 	cmd.Env = os.Environ()
	// }*/

	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if cmdErr := cmd.Start(); cmdErr != nil {
		a.Logger.Debugln(cmdErr)
		return "", cmdErr.Error(), 65, cmdErr
	}
	pid := int32(cmd.Process.Pid)

	// custom context handling, we need to kill child procs if this is a batch script,
	// otherwise it will hang forever
	// the normal exec.CommandContext() doesn't work since it only kills the parent process
	go func(p int32) {
		<-ctx.Done()
		_ = common.KillProc(p)
		timedOut = true
	}(pid)

	cmdErr := cmd.Wait()

	if timedOut {
		stdout = outb.String()
		stderr = fmt.Sprintf("%s\nScript timed out after %d seconds", errb.String(), timeout)
		exitcode = 98
		a.Logger.Debugln("Script check timeout:", ctx.Err())
	} else {
		stdout = outb.String()
		stderr = errb.String()

		// get the exit code
		if cmdErr != nil {
			if exitError, ok := cmdErr.(*exec.ExitError); ok {
				if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
					exitcode = ws.ExitStatus()
				} else {
					exitcode = defaultExitCode
				}
			} else {
				exitcode = defaultExitCode
			}
		} else {
			if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
				exitcode = ws.ExitStatus()
			} else {
				exitcode = 0
			}
		}
	}
	return stdout, stderr, exitcode, nil
}

// ScriptCheck Runs either a batch file, PowerShell or Python script,
// and sends the results back to the server
func (a *windowsAgent) ScriptCheck(data rmm.Check, r *resty.Client) {
	start := time.Now()
	stdout, stderr, retcode, _ := a.RunScript(data.Script.Code, data.Script.Interpreter, data.ScriptArgs, data.Timeout)

	payload := map[string]interface{}{
		"id":      data.CheckPK,
		"stdout":  stdout,
		"stderr":  stderr,
		"retcode": retcode,
		"runtime": time.Since(start).Seconds(),
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

// DiskCheck checks disk usage
func (a *windowsAgent) DiskCheck(data rmm.Check, r *resty.Client) {
	var payload map[string]interface{}

	usage, err := disk.Usage(data.Storage)
	if err != nil {
		a.Logger.Debugln("StorageDrive", data.Storage, err)

		payload = map[string]interface{}{
			"id":     data.CheckPK,
			"exists": false,
		}

		if _, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER); err != nil {
			a.Logger.Debugln(err)
		}
		return
	}

	payload = map[string]interface{}{
		"id":           data.CheckPK,
		"exists":       true,
		"percent_used": usage.UsedPercent,
		"total":        usage.Total,
		"free":         usage.Free,
		// todo: 2021-12-31: "more_info" ?
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

// CPULoadCheck Checks the average processor load
func (a *windowsAgent) CPULoadCheck(data rmm.Check, r *resty.Client) {
	payload := map[string]interface{}{
		"id":      data.CheckPK,
		"percent": a.GetCPULoadAvg(),
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

// MemCheck Checks memory usage percentage
func (a *windowsAgent) MemCheck(data rmm.Check, r *resty.Client) {
	host, _ := ps.Host()
	mem, _ := host.Memory()
	percent := (float64(mem.Used) / float64(mem.Total)) * 100

	payload := map[string]interface{}{
		"id":      data.CheckPK,
		"percent": int(math.Round(percent)),
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

// EventLogCheck Retrieve the Windows Event Logs
func (a *windowsAgent) EventLogCheck(data rmm.Check, r *resty.Client) {
	evtLog := a.GetEventLog(data.LogName, data.SearchLastDays)

	payload := map[string]interface{}{
		"id":  data.CheckPK,
		"log": evtLog,
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

// PingCheck Plays ping pong
func (a *windowsAgent) PingCheck(data rmm.Check, r *resty.Client) {
	cmdArgs := []string{data.IP}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(90)*time.Second)
	defer cancel()

	var (
		outb   bytes.Buffer
		errb   bytes.Buffer
		hasOut bool
		hasErr bool
		output string
	)
	cmd := exec.CommandContext(ctx, "ping", cmdArgs...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	cmdErr := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		a.Logger.Debugln("Ping check:", ctx.Err())
		hasErr = true
		output = fmt.Sprintf("Ping check %s timed out", data.IP)
	} else if cmdErr != nil || errb.String() != "" {
		hasErr = true
		output = fmt.Sprintf("%s\n%s", outb.String(), errb.String())
	} else {
		hasOut = true
		output = outb.String()
	}

	// todo: 2021-12-31: payload structure changed in later versions
	payload := map[string]interface{}{
		"id":         data.CheckPK,
		"has_stdout": hasOut,
		"has_stderr": hasErr,
		"output":     output,
		// todo: 2021-12-31: "status":
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

// CheckService Checks a Windows Service
func (a *windowsAgent) CheckService(data rmm.Check, r *resty.Client) {
	var status string
	exists := true

	status, err := GetServiceStatus(data.ServiceName)
	if err != nil {
		exists = false
		status = "n/a"
		a.Logger.Debugln("Service", data.ServiceName, err)
	}

	payload := map[string]interface{}{
		"id":     data.CheckPK,
		"exists": exists,
		"status": status,
	}

	resp, err := r.R().SetBody(payload).Patch(API_URL_CHECKRUNNER)
	if err != nil {
		a.Logger.Debugln(err)
		return
	}

	a.handleAssignedTasks(resp.String(), data.AssignedTasks)
}

func (a *windowsAgent) handleAssignedTasks(status string, tasks []rmm.AssignedTask) {
	if len(tasks) > 0 && status == "failing" {
		var wg sync.WaitGroup
		for _, t := range tasks {
			if t.Enabled {
				wg.Add(1)
				go func(pk int, wg *sync.WaitGroup) {
					defer wg.Done()
					a.RunTask(pk)
				}(t.TaskPK, &wg)
			}
		}
		wg.Wait()
	}
}
