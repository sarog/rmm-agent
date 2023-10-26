package windows

import (
	"math/rand"
	"sync"
	"time"

	jrmm "github.com/jetrmm/rmm-shared"
	"github.com/nats-io/nats.go"
	rmm "github.com/sarog/rmmagent/shared"
	"github.com/ugorji/go/codec"
)

const (
	// Deprecated since 1.7.0, replaced with NATS
	API_URL_CHECKIN = "/api/v3/checkin/"

	CHECKIN_MODE_DISKS        = "disks"
	CHECKIN_MODE_HELLO        = "hello"
	CHECKIN_MODE_LOGGEDONUSER = "loggedonuser"
	CHECKIN_MODE_OSINFO       = "osinfo"
	CHECKIN_MODE_PUBLICIP     = "publicip"
	CHECKIN_MODE_SOFTWARE     = "software"
	CHECKIN_MODE_STARTUP      = "startup"
	CHECKIN_MODE_WINSERVICES  = "winservices"

	NATS_MODE_DISKS       = "agent-disks"
	NATS_MODE_HELLO       = "agent-hello"
	NATS_MODE_OSINFO      = "agent-agentinfo"
	NATS_MODE_PUBLICIP    = "agent-publicip"
	NATS_MODE_WINSERVICES = "agent-winsvc"
	NATS_MODE_WMI         = "agent-wmi" // sysinfo?
)

func (a *windowsAgent) RunAgentService(nc *nats.Conn) {
	var wg sync.WaitGroup
	wg.Add(1)
	go a.WinAgentSvc(nc)
	go a.CheckRunner()
	wg.Wait()
}

func (a *windowsAgent) WinAgentSvc(nc *nats.Conn) {
	a.Logger.Infoln("Agent service started")

	a.CreateAgentTempDir()

	sleepDelay := randRange(14, 22)
	a.Logger.Debugf("Sleeping for %v seconds", sleepDelay)
	time.Sleep(time.Duration(sleepDelay) * time.Second)

	// a.RunMigrations()

	startup := []string{CHECKIN_MODE_HELLO, CHECKIN_MODE_OSINFO, CHECKIN_MODE_WINSERVICES, CHECKIN_MODE_DISKS, CHECKIN_MODE_PUBLICIP, CHECKIN_MODE_SOFTWARE, CHECKIN_MODE_LOGGEDONUSER}
	for _, s := range startup {
		a.CheckIn(nc, s)
		time.Sleep(time.Duration(randRange(300, 900)) * time.Millisecond)
	}

	time.Sleep(1 * time.Second)
	a.CheckForRecovery()

	time.Sleep(time.Duration(randRange(2, 7)) * time.Second)
	a.CheckIn(nc, CHECKIN_MODE_STARTUP)

	checkInTicker := time.NewTicker(time.Duration(randRange(40, 110)) * time.Second)
	checkInOSTicker := time.NewTicker(time.Duration(randRange(250, 450)) * time.Second)
	checkInWinSvcTicker := time.NewTicker(time.Duration(randRange(700, 1000)) * time.Second)
	checkInPubIPTicker := time.NewTicker(time.Duration(randRange(300, 500)) * time.Second)
	checkInDisksTicker := time.NewTicker(time.Duration(randRange(200, 600)) * time.Second)
	checkInLoggedUserTicker := time.NewTicker(time.Duration(randRange(850, 1400)) * time.Second)
	checkInSWTicker := time.NewTicker(time.Duration(randRange(2400, 3000)) * time.Second)
	recoveryTicker := time.NewTicker(time.Duration(randRange(180, 300)) * time.Second)

	for {
		select {
		case <-checkInTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_HELLO)
		case <-checkInOSTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_OSINFO)
		case <-checkInWinSvcTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_WINSERVICES)
		case <-checkInPubIPTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_PUBLICIP)
		case <-checkInDisksTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_DISKS)
		case <-checkInLoggedUserTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_LOGGEDONUSER)
		case <-checkInSWTicker.C:
			a.CheckIn(nc, CHECKIN_MODE_SOFTWARE)
		case <-recoveryTicker.C:
			a.CheckForRecovery()
		}
	}
}

// CheckIn Check in with the server
func (a *windowsAgent) CheckIn(nc *nats.Conn, mode string) {
	var rerr error
	var payload interface{}
	var nMode string

	// Outgoing payload to server
	switch mode {
	case CHECKIN_MODE_HELLO:
		nMode = NATS_MODE_HELLO
		payload = jrmm.CheckInNats{
			AgentId: a.AgentID,
			Version: a.Version,
		}

	case CHECKIN_MODE_STARTUP:
		// server will then request 2 calls via nats:
		//  'installchoco' and 'getwinupdates'
		payload = rmm.AgentHeader{
			Func:    "startup",
			AgentId: a.AgentID,
			Version: a.Version,
		}

	case CHECKIN_MODE_OSINFO:
		plat, osInfo := a.OSInfo()
		reboot, err := a.SystemRebootRequired()
		if err != nil {
			reboot = false
		}

		nMode = NATS_MODE_OSINFO
		payload = jrmm.AgentInfoNats{
			AgentId:      a.AgentID,
			Username:     a.LoggedOnUser(),
			Hostname:     a.Hostname,
			OS:           osInfo,
			Platform:     plat,
			TotalRAM:     a.TotalRAM(),
			BootTime:     a.BootTime(),
			RebootNeeded: reboot,
		}

	case CHECKIN_MODE_WINSERVICES:
		nMode = NATS_MODE_WINSERVICES
		payload = jrmm.WinSvcNats{
			AgentId: a.AgentID,
			WinSvcs: a.GetServicesNATS(),
		}

	case CHECKIN_MODE_PUBLICIP:
		nMode = NATS_MODE_PUBLICIP
		payload = jrmm.PublicIPNats{
			AgentId:  a.AgentID,
			PublicIP: a.PublicIP(),
		}

	case CHECKIN_MODE_DISKS:
		nMode = NATS_MODE_DISKS
		payload = jrmm.WinDisksNats{
			AgentId: a.AgentID,
			Drives:  a.GetStorage(),
		}

	case CHECKIN_MODE_LOGGEDONUSER:
		payload = rmm.CheckInLoggedUser{
			AgentHeader: rmm.AgentHeader{
				Func:    "loggedonuser",
				AgentId: a.AgentID,
				Version: a.Version,
			},
			Username: a.LoggedOnUser(),
		}

	case CHECKIN_MODE_SOFTWARE:
		payload = rmm.CheckInSW{
			AgentHeader: rmm.AgentHeader{
				Func:    "software",
				AgentId: a.AgentID,
				Version: a.Version,
			},
			InstalledSW: a.GetInstalledSoftware(),
		}
	}

	// todo: 2022-01-02: add error logging/handling
	if len(nMode) > 0 {
		// opts := a.SetupNatsOptions()
		// server := fmt.Sprintf("tls://%s:%d", a.ApiURL, a.ApiPort)
		// nc, err := nats.Connect(server, opts...)
		// if err != nil {
		// 	a.Logger.Errorln(err)
		// } else {
		var response []byte
		err := codec.NewEncoderBytes(&response, new(codec.MsgpackHandle)).Encode(payload)
		if err != nil {
			return
		}
		nc.PublishRequest(a.AgentID, nMode, response)
		// was testing with: nc.Publish(a.AgentID, cPayload)
		// }
		// mh.RawToString = true
		// dec := codec.NewDecoderBytes(msg.Data, &mh)
		// if err := dec.Decode(&payload); err != nil {
		// 	a.Logger.Errorln(err)
		// 	return
		// }
		// nc.Flush()
		// nc.Close()
	} else {
		// Deprecated endpoint
		if mode == CHECKIN_MODE_HELLO {
			// _, rerr = a.RClient.R().SetBody(payload).Patch(API_URL_CHECKIN)
			// a.CheckIn(CHECKIN_MODE_HELLO)
			// time.Sleep(200 * time.Millisecond)
		} else if mode == CHECKIN_MODE_STARTUP {
			_, rerr = a.RClient.R().SetBody(payload).Post(API_URL_CHECKIN)
		} else {
			// 'put' is deprecated as of 1.7.0
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
