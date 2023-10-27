package windows

import (
	"fmt"
	ps "github.com/jetrmm/go-sysinfo"
	gops "github.com/shirou/gopsutil/v3/process"
)

type ProcessMsg struct {
	Name     string `json:"name"`
	Pid      int    `json:"pid"`
	MemBytes uint64 `json:"membytes"`
	Username string `json:"username"`
	UID      int    `json:"id"`
	CPU      string `json:"cpu_percent"`
}

func (a *windowsAgent) GetProcsRPC() []ProcessMsg {
	ret := make([]ProcessMsg, 0)

	procs, _ := ps.Processes()
	for i, process := range procs {
		p, err := process.Info()
		if err != nil {
			continue
		}
		if p.PID == 0 {
			continue
		}

		m, _ := process.Memory()
		proc, gerr := gops.NewProcess(int32(p.PID))
		if gerr != nil {
			continue
		}
		cpu, _ := proc.CPUPercent()
		user, _ := proc.Username()

		ret = append(ret, ProcessMsg{
			Name:     p.Name,
			Pid:      p.PID,
			MemBytes: m.Resident,
			Username: user,
			UID:      i,
			CPU:      fmt.Sprintf("%.1f", cpu),
		})
	}
	return ret
}

// ChecksRunning prevents duplicate checks from running
// Have to do it this way, can't use atomic because they can run from both rpc and rmmagent services
func (a *windowsAgent) ChecksRunning() bool {
	running := false
	procs, err := ps.Processes()
	if err != nil {
		return running
	}

Out:
	for _, process := range procs {
		p, err := process.Info()
		if err != nil {
			continue
		}
		if p.PID == 0 {
			continue
		}
		if p.Exe != a.AgentExe {
			continue
		}

		for _, arg := range p.Args {
			if arg == "runchecks" || arg == "checkrunner" {
				running = true
				break Out
			}
		}
	}
	return running
}
