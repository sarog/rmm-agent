package windows

import (
	"fmt"
	"github.com/sarog/rmmagent/agent/common"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/ugorji/go/codec"
)

type NatsMsg struct {
	Func            string            `json:"func"`
	Timeout         int               `json:"timeout"`
	Data            map[string]string `json:"payload"`
	ScriptArgs      []string          `json:"script_args"`
	ProcPID         int32             `json:"procpid"`
	TaskPK          int               `json:"taskpk"`
	ScheduledTask   SchedTask         `json:"schedtaskpayload"`
	RecoveryCommand string            `json:"recoverycommand"`
	UpdateGUIDs     []string          `json:"guids"`           // todo: move
	ChocoProgName   string            `json:"choco_prog_name"` // todo: move
	PendingActionPK int               `json:"pending_action_pk"`
}

var (
	agentUpdateLocker      uint32
	getWinUpdateLocker     uint32
	installWinUpdateLocker uint32
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

// RunService handles incoming NATS payloads from server
func (a *windowsAgent) RunService() {
	a.Logger.Infoln("RPC service started")
	opts := a.SetupNatsOptions()
	server := fmt.Sprintf("tls://%s:%d", a.ApiURL, a.ApiPort)
	nc, err := nats.Connect(server, opts...)
	if err != nil {
		a.Logger.Fatalln(err)
	}

	go a.RunAgentService(nc)
	var wg sync.WaitGroup
	wg.Add(1)

	// todo: 2023-10-17: JetStream
	// Migration: https://natsbyexample.com/examples/jetstream/api-migration/go
	// https://github.com/nats-io/nats.go#jetstream
	// js, _ := jetstream.New(nc)

	// Incoming payload from server
	nc.Subscribe(a.AgentID, func(msg *nats.Msg) {
		a.Logger.SetOutput(os.Stdout)
		var payload *NatsMsg
		var mh codec.MsgpackHandle
		mh.RawToString = true

		dec := codec.NewDecoderBytes(msg.Data, &mh)
		if err := dec.Decode(&payload); err != nil {
			a.Logger.Errorln(err)
			return
		}

		switch payload.Func {
		case NATS_CMD_PING:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				a.Logger.Debugln("pong")
				ret.Encode("pong")
				msg.Respond(resp)
			}()

		case NATS_CMD_TASK_ADD:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				success, err := a.CreateSchedTask(p.ScheduledTask)
				if err != nil {
					a.Logger.Errorln(err.Error())
					ret.Encode(err.Error())
				} else if !success {
					ret.Encode("Something went wrong")
				} else {
					ret.Encode("ok")
				}
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_TASK_DEL:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				err := DeleteSchedTask(p.ScheduledTask.Name)
				if err != nil {
					a.Logger.Errorln(err.Error())
					ret.Encode(err.Error())
				} else {
					ret.Encode("ok")
				}
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_TASK_ENABLE:
			//  1.7.3+: replaced with 'func: schedtask': (modify_task_on_agent)
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				err := EnableSchedTask(p.ScheduledTask)
				if err != nil {
					a.Logger.Errorln(err.Error())
					ret.Encode(err.Error())
				} else {
					ret.Encode("ok")
				}
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_TASK_LIST:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				tasks := ListSchedTasks()
				a.Logger.Debugln(tasks)
				ret.Encode(tasks)
				msg.Respond(resp)
			}()

		case NATS_CMD_EVENTLOG:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				days, _ := strconv.Atoi(p.Data["days"])
				evtLog := a.GetEventLog(p.Data["logname"], days)
				a.Logger.Debugln(evtLog)
				ret.Encode(evtLog)
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_PROCS_LIST:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				procs := a.GetProcsRPC()
				a.Logger.Debugln(procs)
				ret.Encode(procs)
				msg.Respond(resp)
			}()

		case NATS_CMD_PROCS_KILL:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				err := common.KillProc(p.ProcPID)
				if err != nil {
					ret.Encode(err.Error())
					a.Logger.Debugln(err.Error())
				} else {
					ret.Encode("ok")
				}
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_RAWCMD:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				out, _ := CMDShell(p.Data["shell"], []string{}, p.Data["command"], p.Timeout, false)
				a.Logger.Debugln(out)
				if out[1] != "" {
					ret.Encode(out[1])
				} else {
					ret.Encode(out[0])
				}

				msg.Respond(resp)
			}(payload)

		case NATS_CMD_WINSERVICES:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				svcs := a.GetServices()
				a.Logger.Debugln(svcs)
				ret.Encode(svcs)
				msg.Respond(resp)
			}()

		case NATS_CMD_WINSVC_DETAIL:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				svc := a.GetServiceDetail(p.Data["name"])
				a.Logger.Debugln(svc)
				ret.Encode(svc)
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_WINSVC_ACTION:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				retData := a.ControlService(p.Data["name"], p.Data["action"])
				a.Logger.Debugln(retData)
				ret.Encode(retData)
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_WINSVC_EDIT:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				retData := a.EditService(p.Data["name"], p.Data["startType"])
				a.Logger.Debugln(retData)
				ret.Encode(retData)
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_SCRIPT_RUN:
			go func(p *NatsMsg) {
				var resp []byte
				var retData string
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				stdout, stderr, _, err := a.RunScript(p.Data["code"], p.Data["shell"], p.ScriptArgs, p.Timeout)
				if err != nil {
					a.Logger.Debugln(err)
					retData = err.Error()
				} else {
					retData = stdout + stderr
				}
				a.Logger.Debugln(retData)
				ret.Encode(retData)
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_SCRIPT_RUN_FULL:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				start := time.Now()
				out, err, retcode, _ := a.RunScript(p.Data["code"], p.Data["shell"], p.ScriptArgs, p.Timeout)
				retData := struct {
					Stdout   string  `json:"stdout"`
					Stderr   string  `json:"stderr"`
					Retcode  int     `json:"retcode"`
					ExecTime float64 `json:"execution_time"`
				}{out, err, retcode, time.Since(start).Seconds()}
				a.Logger.Debugln(retData)
				ret.Encode(retData)
				msg.Respond(resp)
			}(payload)

		case NATS_CMD_RECOVER:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))

				switch p.Data["mode"] {
				case "jetagent":
					a.Logger.Debugln("Recovering agent")
					a.RecoverAgent()
				}

				ret.Encode("ok")
				msg.Respond(resp)
			}(payload)

		case "recoverycmd": // 2022-01-01: removed or merged
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				ret.Encode("ok")
				msg.Respond(resp)
				a.RecoverCMD(p.RecoveryCommand)
			}(payload)

		case NATS_CMD_SOFTWARE_LIST:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				sw := a.GetInstalledSoftware()
				a.Logger.Debugln(sw)
				ret.Encode(sw)
				msg.Respond(resp)
			}()

		case NATS_CMD_REBOOT_NOW:
			go func() {
				a.Logger.Debugln("Scheduling immediate reboot")
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				ret.Encode("ok")
				msg.Respond(resp)
				_, _ = CMD("shutdown.exe", []string{"/r", "/t", "5", "/f"}, 15, false)
			}()

		case NATS_CMD_REBOOT_NEEDED: // 2022-01-01: removed or merged
			go func() {
				a.Logger.Debugln("Checking if a reboot is needed")
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				out, err := a.SystemRebootRequired()
				if err == nil {
					a.Logger.Debugln("Reboot needed:", out)
					ret.Encode(out)
				} else {
					a.Logger.Debugln("Error checking if a reboot is needed:", err)
					ret.Encode(false)
				}
				msg.Respond(resp)
			}()

		case NATS_CMD_SYSINFO:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				a.Logger.Debugln("Getting system info via WMI")

				modes := []string{CHECKIN_MODE_OSINFO, CHECKIN_MODE_PUBLICIP, CHECKIN_MODE_DISKS}
				for _, mode := range modes {
					a.CheckIn(nc, mode)
					time.Sleep(200 * time.Millisecond)
				}
				a.SysInfo()
				ret.Encode("ok")
				msg.Respond(resp)
			}()

		case NATS_CMD_SYNC:
			go func() {
				a.Logger.Debugln("Sending system info and software")
				a.SyncInfo()
			}()

		case NATS_CMD_WMI:
			go func() {
				a.Logger.Debugln("Sending WMI")
				a.SysInfo()
			}()

		case NATS_CMD_CPULOADAVG:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				a.Logger.Debugln("Getting CPU load average")
				loadAvg := a.GetCPULoadAvg()
				a.Logger.Debugln("CPU load average:", loadAvg)
				ret.Encode(loadAvg)
				msg.Respond(resp)
			}()

		case NATS_CMD_RUNCHECKS:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				if a.ChecksRunning() {
					ret.Encode("busy")
					msg.Respond(resp)
					a.Logger.Debugln("Checks are already running, please wait")
				} else {
					ret.Encode("ok")
					msg.Respond(resp)
					a.Logger.Debugln("Running checks")
					// todo: verify:
					_, checkerr := CMD(a.AgentExe, []string{"-m", "runchecks"}, 600, false)
					if checkerr != nil {
						a.Logger.Errorln("RPC RunChecks", checkerr)
					}
				}
			}()

		case NATS_CMD_TASK_RUN:
			go func(p *NatsMsg) {
				a.Logger.Debugln("Running task")
				a.RunTask(p.TaskPK)
			}(payload)

		case NATS_CMD_PUBLICIP:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				ret.Encode(a.PublicIP())
				msg.Respond(resp)
			}()

		case NATS_CMD_INSTALL_CHOCO:
			go a.InstallChoco()

		case NATS_CMD_CHOCO_INSTALL:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				ret.Encode("ok")
				msg.Respond(resp)
				out, _ := a.InstallWithChoco(p.ChocoProgName)
				results := map[string]string{"results": out}
				url := fmt.Sprintf("/api/v3/%d/chocoresult/", p.PendingActionPK)
				a.RClient.R().SetBody(results).Patch(url)
			}(payload)

		case NATS_CMD_GETWINUPDATES:
			go func() {
				if !atomic.CompareAndSwapUint32(&getWinUpdateLocker, 0, 1) {
					a.Logger.Debugln("Already checking for Windows Updates")
				} else {
					a.Logger.Debugln("Checking for Windows Updates")
					defer atomic.StoreUint32(&getWinUpdateLocker, 0)
					a.GetWinUpdates()
				}
			}()

		case NATS_CMD_INSTALL_WINUPDATES:
			go func(p *NatsMsg) {
				if !atomic.CompareAndSwapUint32(&installWinUpdateLocker, 0, 1) {
					a.Logger.Debugln("Already installing Windows Updates")
				} else {
					a.Logger.Debugln("Installing Windows Updates", p.UpdateGUIDs)
					defer atomic.StoreUint32(&installWinUpdateLocker, 0)
					a.InstallUpdates(p.UpdateGUIDs)
				}
			}(payload)

		case NATS_CMD_AGENT_UPDATE:
			go func(p *NatsMsg) {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				if !atomic.CompareAndSwapUint32(&agentUpdateLocker, 0, 1) {
					a.Logger.Debugln("Agent update already running")
					ret.Encode("updaterunning") // todo: 2022-01-02: removed or renamed? no mention on server side
					msg.Respond(resp)
				} else {
					ret.Encode("ok")
					msg.Respond(resp)
					a.AgentUpdate(p.Data["url"], p.Data["inno"], p.Data["version"])
					atomic.StoreUint32(&agentUpdateLocker, 0)
					nc.Flush()
					nc.Close()
					os.Exit(0)
				}
			}(payload)

		case NATS_CMD_AGENT_UNINSTALL:
			go func() {
				var resp []byte
				ret := codec.NewEncoderBytes(&resp, new(codec.MsgpackHandle))
				ret.Encode("ok")
				msg.Respond(resp)
				a.AgentUninstall()
				nc.Flush()
				nc.Close()
				os.Exit(0)
			}()
		}
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		a.Logger.Errorln(err)
		os.Exit(1)
	}

	runtime.Goexit()
}
