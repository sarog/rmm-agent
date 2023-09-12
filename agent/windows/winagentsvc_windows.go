package windows

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	rmm "github.com/sarog/rmmagent/shared"
	trmm "github.com/sarog/trmm-shared"
	"github.com/ugorji/go/codec"
)

const (
	// Deprecated
	// Since 1.7.0, replaced with NATS
	API_URL_CHECKIN = "/api/v3/checkin/"

	// CheckIn
	CHECKIN_MODE_DISKS        = "disks"
	CHECKIN_MODE_HELLO        = "hello"
	CHECKIN_MODE_LOGGEDONUSER = "loggedonuser"
	CHECKIN_MODE_OSINFO       = "osinfo"
	CHECKIN_MODE_PUBLICIP     = "publicip"
	CHECKIN_MODE_SOFTWARE     = "software"
	CHECKIN_MODE_STARTUP      = "startup"
	CHECKIN_MODE_WINSERVICES  = "winservices"

	// nats service: natsapi/svc.go:16
	NATS_MODE_DISKS       = "agent-disks"
	NATS_MODE_HELLO       = "agent-hello"
	NATS_MODE_OSINFO      = "agent-agentinfo"
	NATS_MODE_PUBLICIP    = "agent-publicip"
	NATS_MODE_WINSERVICES = "agent-winsvc"
	NATS_MODE_WMI         = "agent-wmi" // sysinfo?
)

// RunAgentService
func (a *WindowsAgent) RunAgentService() {
	var wg sync.WaitGroup
	wg.Add(1)
	go a.WinAgentSvc()
	go a.CheckRunner()
	wg.Wait()
}

// WinAgentSvc Windows nssm service startup
func (a *WindowsAgent) WinAgentSvc() {
	a.Logger.Infoln("Agent service started")

	// go a.GetPython(false)

	a.CreateAgentTempDir()

	sleepDelay := randRange(14, 22)
	a.Logger.Debugf("Sleeping for %v seconds", sleepDelay)
	time.Sleep(time.Duration(sleepDelay) * time.Second)

	a.RunMigrations()

	startup := []string{CHECKIN_MODE_HELLO, CHECKIN_MODE_OSINFO, CHECKIN_MODE_WINSERVICES, CHECKIN_MODE_DISKS, CHECKIN_MODE_PUBLICIP, CHECKIN_MODE_SOFTWARE, CHECKIN_MODE_LOGGEDONUSER}
	for _, s := range startup {
		a.CheckIn(s)
		time.Sleep(time.Duration(randRange(300, 900)) * time.Millisecond)
	}
	a.SyncMeshNodeID()
	time.Sleep(1 * time.Second)
	a.CheckForRecovery()

	time.Sleep(time.Duration(randRange(2, 7)) * time.Second)
	a.CheckIn(CHECKIN_MODE_STARTUP)

	checkInTicker := time.NewTicker(time.Duration(randRange(40, 110)) * time.Second)
	checkInOSTicker := time.NewTicker(time.Duration(randRange(250, 450)) * time.Second)
	checkInWinSvcTicker := time.NewTicker(time.Duration(randRange(700, 1000)) * time.Second)
	checkInPubIPTicker := time.NewTicker(time.Duration(randRange(300, 500)) * time.Second)
	checkInDisksTicker := time.NewTicker(time.Duration(randRange(200, 600)) * time.Second)
	checkInLoggedUserTicker := time.NewTicker(time.Duration(randRange(850, 1400)) * time.Second)
	checkInSWTicker := time.NewTicker(time.Duration(randRange(2400, 3000)) * time.Second)
	syncMeshTicker := time.NewTicker(time.Duration(randRange(2400, 2900)) * time.Second)
	recoveryTicker := time.NewTicker(time.Duration(randRange(180, 300)) * time.Second)

	for {
		select {
		case <-checkInTicker.C:
			a.CheckIn(CHECKIN_MODE_HELLO)
		case <-checkInOSTicker.C:
			a.CheckIn(CHECKIN_MODE_OSINFO)
		case <-checkInWinSvcTicker.C:
			a.CheckIn(CHECKIN_MODE_WINSERVICES)
		case <-checkInPubIPTicker.C:
			a.CheckIn(CHECKIN_MODE_PUBLICIP)
		case <-checkInDisksTicker.C:
			a.CheckIn(CHECKIN_MODE_DISKS)
		case <-checkInLoggedUserTicker.C:
			a.CheckIn(CHECKIN_MODE_LOGGEDONUSER)
		case <-checkInSWTicker.C:
			a.CheckIn(CHECKIN_MODE_SOFTWARE)
		case <-syncMeshTicker.C:
			a.SyncMeshNodeID()
		case <-recoveryTicker.C:
			a.CheckForRecovery()
		}
	}
}

// CheckIn Check in with the server
func (a *WindowsAgent) CheckIn(mode string) {
	var rerr error
	var payload interface{}
	var nPayload interface{}
	var nMode string

	// Outgoing payload to server
	switch mode {
	case CHECKIN_MODE_HELLO:
		// 2022-01-01: 'agent-hello' @ natsapi/svc.go:36
		nMode = NATS_MODE_HELLO
		nPayload = trmm.CheckInNats{
			Agentid: a.AgentID(),
			Version: a.Version,
		}

		/*payload = rmm.CheckIn{
			Func:    "hello",
			Agentid: a.AgentID(),
			Version: a.Version,
		}*/

	case CHECKIN_MODE_STARTUP:
		// 2022-01-01: relies on 'post' endpoint @ api/tacticalrmm/apiv3/views.py:84
		// server will then request 2 calls via nats:
		//  'installchoco' @ api/tacticalrmm/apiv3/views.py:87
		// 	'getwinupdates' @ api/tacticalrmm/apiv3/views.py:90
		payload = rmm.CheckIn{
			Func:    "startup",
			Agentid: a.AgentID(),
			Version: a.Version,
		}

	case CHECKIN_MODE_OSINFO:
		plat, osInfo := a.OSInfo()
		reboot, err := a.SystemRebootRequired()
		if err != nil {
			reboot = false
		}

		// 2022-01-01: 'agent-agentinfo' @ natsapi/svc.go:70
		nMode = NATS_MODE_OSINFO
		nPayload = trmm.AgentInfoNats{
			Agentid:      a.AgentID(),
			Username:     a.LoggedOnUser(),
			Hostname:     a.Hostname(),
			OS:           osInfo,
			Platform:     plat,
			TotalRAM:     a.TotalRAM(),
			BootTime:     a.BootTime(),
			RebootNeeded: reboot,
		}

		/*payload = rmm.CheckInOS{
			CheckIn: rmm.CheckIn{
				Func:    "osinfo",
				Agentid: a.AgentID(),
				Version: a.Version,
			},
			Hostname:     a.Hostname,
			OS:           osInfo,
			Platform:     plat,
			TotalRAM:     a.TotalRAM(),
			BootTime:     a.BootTime(),
			RebootNeeded: reboot,
		}*/

	case CHECKIN_MODE_WINSERVICES:
		// 2022-01-01: 'agent-winsvc' @ natsapi/svc.go:117
		nMode = NATS_MODE_WINSERVICES
		nPayload = trmm.WinSvcNats{
			Agentid: a.AgentID(),
			WinSvcs: a.GetServicesNATS(),
		}

		/*payload = rmm.CheckInWinServices{
			CheckIn: rmm.CheckIn{
				Func:    "winservices",
				Agentid: a.AgentID(),
				Version: a.Version,
			},
			Services: a.GetServices(),
		}*/

	case CHECKIN_MODE_PUBLICIP:
		// 2022-01-01: 'agent-publicip' @ natsapi/svc.go:56
		nMode = NATS_MODE_PUBLICIP
		nPayload = trmm.PublicIPNats{
			Agentid:  a.AgentID(),
			PublicIP: a.PublicIP(),
		}

		/*payload = rmm.CheckInPublicIP{
			CheckIn: rmm.CheckIn{
				Func:    "publicip",
				Agentid: a.AgentID(),
				Version: a.Version,
			},
			PublicIP: a.PublicIP(),
		}*/

	case CHECKIN_MODE_DISKS:
		// 2022-01-01: 'agent-disks' @ natsapi/svc.go:97
		nMode = NATS_MODE_DISKS
		nPayload = trmm.WinDisksNats{
			Agentid: a.AgentID(),
			Disks:   a.GetDisksNATS(),
		}

		/*payload = rmm.CheckInDisk{
			CheckIn: rmm.CheckIn{
				Func:    "disks",
				Agentid: a.AgentID(),
				Version: a.Version,
			},
			Disks: a.GetDisks(),
		}*/

	case CHECKIN_MODE_LOGGEDONUSER:
		// 2022-01-02: this is combined with 'agent-hello'
		// 2022-01-01: api/tacticalrmm/apiv3/views.py:61
		payload = rmm.CheckInLoggedUser{
			CheckIn: rmm.CheckIn{
				Func:    "loggedonuser",
				Agentid: a.AgentID(),
				Version: a.Version,
			},
			Username: a.LoggedOnUser(),
		}

	case CHECKIN_MODE_SOFTWARE:
		// 2022-01-01: api/tacticalrmm/apiv3/views.py:67
		payload = rmm.CheckInSW{
			CheckIn: rmm.CheckIn{
				Func:    "software",
				Agentid: a.AgentID(),
				Version: a.Version,
			},
			InstalledSW: a.GetInstalledSoftware(),
		}
	}

	// 2022-01-02
	// todo: 2022-01-02: add error logging/handling
	if len(nMode) > 0 {
		opts := a.setupNatsOptions()
		server := fmt.Sprintf("tls://%s:%d", a.ApiURL, a.ApiPort)
		nc, err := nats.Connect(server, opts...)
		if err != nil {
			a.Logger.Errorln(err)
		} else {
			var cPayload []byte
			codec.NewEncoderBytes(&cPayload, new(codec.MsgpackHandle)).Encode(nPayload)
			// todo: 2022-01-02: test
			err = nc.PublishRequest(a.AgentID(), nMode, cPayload)
			// was testing with: nc.Publish(a.AgentID(), cPayload)
		}
		// mh.RawToString = true
		// dec := codec.NewDecoderBytes(msg.Data, &mh)
		// if err := dec.Decode(&payload); err != nil {
		// 	a.Logger.Errorln(err)
		// 	return
		// }
		nc.Flush()
		nc.Close()
	} else {
		// Deprecated endpoint
		if mode == CHECKIN_MODE_HELLO {
			// 2022-01-01: 'patch' was removed, endpoint deprecated
			// _, rerr = a.RClient.R().SetBody(payload).Patch(API_URL_CHECKIN)
			// a.CheckIn(CHECKIN_MODE_HELLO)
			// time.Sleep(200 * time.Millisecond)
		} else if mode == CHECKIN_MODE_STARTUP {
			// 2022-01-01: api/tacticalrmm/apiv3/views.py:84
			_, rerr = a.RClient.R().SetBody(payload).Post(API_URL_CHECKIN)
		} else {
			// 'put' is deprecated as of 1.7.0
			// 2021-12-31: api/tacticalrmm/apiv3/views.py:30
			_, rerr = a.RClient.R().SetBody(payload).Put(API_URL_CHECKIN)
		}
		if rerr != nil {
			a.Logger.Debugln("Checkin:", rerr)
		}
	}
}

func randRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}
