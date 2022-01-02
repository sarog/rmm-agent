package agent

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gonutz/w32/v2"
	nats "github.com/nats-io/nats.go"
	"golang.org/x/sys/windows/registry"
)

type Installer struct {
	Headers     map[string]string
	RMM         string // API URL
	ClientID    int
	SiteID      int
	Description string
	AgentType   string
	Power       bool
	RDP         bool
	Ping        bool
	WinDefender bool // 2022-01-01: new
	Token       string
	LocalMesh   string
	Cert        string
	Timeout     time.Duration
	SaltMaster  string
	Silent      bool
}

// todo: 2021-12-31: custom branding
// todo: 2022-01-01: perhaps consolidate these elsewhere?
const (
	SERVICE_NAME_RPC        = "tacticalrpc"
	SERVICE_NAME_AGENT      = "tacticalagent"
	SERVICE_NAME_MESHAGENT  = "mesh agent"
	SERVICE_NAME_SALTMINION = "salt-minion"
	SERVICE_DESC_RPC        = "Tactical RMM RPC Service"
	SERVICE_DESC_AGENT      = "Tactical RMM Agent"

	REG_RMM_BASEURL = "BaseURL"
	REG_RMM_AGENTID = "AgentID"
	REG_RMM_APIURL  = "ApiURL"
	REG_RMM_TOKEN   = "Token"
	REG_RMM_AGENTPK = "AgentPK"
	REG_RMM_CERT    = "Cert"
)

func createRegKeys(baseurl, agentid, apiurl, token, agentpk, cert string) {
	// todo: 2021-12-31: migrate to DPAPI?
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\TacticalRMM`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatalln("Error creating registry key:", err)
	}
	defer k.Close()

	err = k.SetStringValue(REG_RMM_BASEURL, baseurl)
	if err != nil {
		log.Fatalln("Error creating BaseURL registry key:", err)
	}

	err = k.SetStringValue(REG_RMM_AGENTID, agentid)
	if err != nil {
		log.Fatalln("Error creating AgentID registry key:", err)
	}

	err = k.SetStringValue(REG_RMM_APIURL, apiurl)
	if err != nil {
		log.Fatalln("Error creating ApiURL registry key:", err)
	}

	err = k.SetStringValue(REG_RMM_TOKEN, token)
	if err != nil {
		log.Fatalln("Error creating Token registry key:", err)
	}

	err = k.SetStringValue(REG_RMM_AGENTPK, agentpk)
	if err != nil {
		log.Fatalln("Error creating AgentPK registry key:", err)
	}

	if len(cert) > 0 {
		err = k.SetStringValue(REG_RMM_CERT, cert)
		if err != nil {
			log.Fatalln("Error creating Cert registry key:", err)
		}
	}
}

func (a *Agent) Install(i *Installer) {
	a.checkExistingAndRemove(i.Silent)

	i.Headers = map[string]string{
		"content-type":  "application/json",
		"Authorization": fmt.Sprintf("Token %s", i.Token),
	}
	a.AgentID = GenerateAgentID()
	a.Logger.Debugln("Agent ID:", a.AgentID)

	u, err := url.Parse(i.RMM)
	if err != nil {
		a.installerMsg(err.Error(), "error", i.Silent)
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		a.installerMsg("Invalid URL (must contain https or http)", "error", i.Silent)
	}

	// This will match either IPv4 or IPv4:port
	var ipPort = regexp.MustCompile(`[0-9]+(?:\.[0-9]+){3}(:[0-9]+)?`)

	// if ipv4:port, strip the port to get ip for salt master
	if ipPort.MatchString(u.Host) && strings.Contains(u.Host, ":") {
		i.SaltMaster = strings.Split(u.Host, ":")[0]
	} else if strings.Contains(u.Host, ":") {
		i.SaltMaster = strings.Split(u.Host, ":")[0]
	} else {
		i.SaltMaster = u.Host
	}

	a.Logger.Debugln("Salt Master:", i.SaltMaster)

	terr := TestTCP(fmt.Sprintf("%s:4222", i.SaltMaster))
	if terr != nil {
		a.installerMsg(fmt.Sprintf("ERROR: Either port 4222 TCP is not open on your RMM, or nats.service is not running.\n\n%s", terr.Error()), "error", i.Silent)
	}

	baseURL := u.Scheme + "://" + u.Host
	a.Logger.Debugln("Base URL:", baseURL)

	iClient := resty.New()
	iClient.SetCloseConnection(true)
	iClient.SetTimeout(15 * time.Second)
	iClient.SetDebug(a.Debug)
	iClient.SetHeaders(i.Headers)
	// 2021-12-31: api/tacticalrmm/apiv3/views.py:475
	creds, cerr := iClient.R().Get(fmt.Sprintf("%s/api/v3/installer/", baseURL))
	if cerr != nil {
		a.installerMsg(cerr.Error(), "error", i.Silent)
	}
	if creds.StatusCode() == 401 {
		a.installerMsg("Installer token has expired. Please generate a new one.", "error", i.Silent)
	}

	// 2021-12-31: api/tacticalrmm/apiv3/views.py:474
	verPayload := map[string]string{"version": a.Version}
	// 2021-12-31: api/tacticalrmm/apiv3/views.py:479
	iVersion, ierr := iClient.R().SetBody(verPayload).Post(fmt.Sprintf("%s/api/v3/installer/", baseURL))
	if ierr != nil {
		a.installerMsg(ierr.Error(), "error", i.Silent)
	}
	if iVersion.StatusCode() != 200 {
		a.installerMsg(DjangoStringResp(iVersion.String()), "error", i.Silent)
	}

	rClient := resty.New()
	rClient.SetCloseConnection(true)
	rClient.SetTimeout(i.Timeout * time.Second)
	rClient.SetDebug(a.Debug)
	// Set REST knox headers
	rClient.SetHeaders(i.Headers)

	// Set local cert if applicable
	if len(i.Cert) > 0 {
		if !FileExists(i.Cert) {
			a.installerMsg(fmt.Sprintf("%s does not exist", i.Cert), "error", i.Silent)
		}
		rClient.SetRootCertificate(i.Cert)
	}

	var arch string
	switch a.Arch {
	case "x86_64":
		arch = "64"
	case "x86":
		arch = "32"
	}

	// Download or copy the mesh-agent.exe
	mesh := filepath.Join(a.ProgramDir, a.MeshInstaller)
	if i.LocalMesh == "" {
		a.Logger.Infoln("Downloading Mesh Agent...")
		payload := map[string]string{"arch": arch}
		// 2022-01-01: api/tacticalrmm/apiv3/views.py:373
		r, err := rClient.R().SetBody(payload).SetOutput(mesh).Post(fmt.Sprintf("%s/api/v3/meshexe/", baseURL))
		if err != nil {
			a.installerMsg(fmt.Sprintf("Failed to download Mesh Agent: %s", err.Error()), "error", i.Silent)
		}
		if r.StatusCode() != 200 {
			a.installerMsg(fmt.Sprintf("Unable to download the Mesh Agent from the RMM server. %s", r.String()), "error", i.Silent)
		}
	} else {
		err := copyFile(i.LocalMesh, mesh)
		if err != nil {
			a.installerMsg(err.Error(), "error", i.Silent)
		}
	}

	a.Logger.Infoln("Installing Mesh Agent...")
	a.Logger.Debugln("Mesh Agent:", mesh)
	meshOut, meshErr := CMD(mesh, []string{"-fullinstall"}, int(90), false)
	if meshErr != nil {
		fmt.Println(meshOut[0])
		fmt.Println(meshOut[1])
		fmt.Println(meshErr)
	}

	fmt.Println(meshOut)
	a.Logger.Debugln("Sleeping for 5 seconds")
	time.Sleep(5 * time.Second)

	meshSuccess := false
	var meshNodeID string
	for !meshSuccess {
		a.Logger.Debugln("Getting Mesh Node ID")
		pMesh, pErr := CMD(a.MeshSystemEXE, []string{"-nodeid"}, int(30), false)
		if pErr != nil {
			a.Logger.Errorln(pErr)
			time.Sleep(5 * time.Second)
			continue
		}
		if pMesh[1] != "" {
			a.Logger.Errorln(pMesh[1])
			time.Sleep(5 * time.Second)
			continue
		}
		meshNodeID = StripAll(pMesh[0])
		a.Logger.Debugln("Node ID:", meshNodeID)
		if strings.Contains(strings.ToLower(meshNodeID), "not defined") {
			a.Logger.Errorln(meshNodeID)
			time.Sleep(5 * time.Second)
			continue
		}
		meshSuccess = true
	}

	a.Logger.Infoln("Adding agent to the dashboard")

	// 2021-12-31: api/tacticalrmm/apiv3/views.py:448
	type NewAgentResp struct {
		AgentPK int    `json:"pk"`
		SaltID  string `json:"saltid"`
		Token   string `json:"token"`
	}

	// 2021-12-31: api/tacticalrmm/apiv3/views.py:409
	agentPayload := map[string]interface{}{
		"agent_id":        a.AgentID,
		"hostname":        a.Hostname,
		"client":          i.ClientID,
		"site":            i.SiteID,
		"mesh_node_id":    meshNodeID,
		"description":     i.Description,
		"monitoring_type": i.AgentType,
	}

	// 2022-01-01: api/tacticalrmm/apiv3/views.py:398
	r, err := rClient.R().SetBody(agentPayload).SetResult(&NewAgentResp{}).Post(fmt.Sprintf("%s/api/v3/newagent/", baseURL))
	if err != nil {
		a.installerMsg(err.Error(), "error", i.Silent)
	}
	if r.StatusCode() != 200 {
		a.installerMsg(r.String(), "error", i.Silent)
	}

	agentPK := r.Result().(*NewAgentResp).AgentPK
	saltID := r.Result().(*NewAgentResp).SaltID
	agentToken := r.Result().(*NewAgentResp).Token

	a.Logger.Debugln("Agent Token:", agentToken)
	a.Logger.Debugln("Agent PK:", agentPK)
	a.Logger.Debugln("Salt ID:", saltID)

	createRegKeys(baseURL, a.AgentID, i.SaltMaster, agentToken, strconv.Itoa(agentPK), i.Cert)
	// Refresh our agent with new values
	a = New(a.Logger, a.Version)

	// Set new headers. No longer knox auth; use agent auth
	rClient.SetHeaders(a.Headers)

	// Send WMI system information
	a.Logger.Debugln("Getting system information with WMI")
	a.GetWMI()

	// Check in once via nats
	opts := a.setupNatsOptions()
	server := fmt.Sprintf("tls://%s:4222", a.ApiURL)

	nc, err := nats.Connect(server, opts...)
	if err != nil {
		a.Logger.Errorln(err)
	} else {
		startup := []string{CHECKIN_MODE_HELLO, CHECKIN_MODE_OSINFO, CHECKIN_MODE_WINSERVICES, CHECKIN_MODE_DISKS, CHECKIN_MODE_PUBLICIP, CHECKIN_MODE_SOFTWARE, CHECKIN_MODE_LOGGEDONUSER}
		for _, mode := range startup {
			a.CheckIn(mode)
			time.Sleep(200 * time.Millisecond)
		}
		nc.Close()
	}

	a.Logger.Debugln("Creating temporary directory")
	a.CreateAgentTempDir()

	a.Logger.Infoln("Installing services...")

	svcCommands := [10][]string{
		// tacticalrpc
		{"install", SERVICE_NAME_RPC, a.EXE, "-m", "rpc"},
		{"set", SERVICE_NAME_RPC, "DisplayName", SERVICE_DESC_RPC},
		{"set", SERVICE_NAME_RPC, "Description", SERVICE_DESC_RPC},
		{"set", SERVICE_NAME_RPC, "AppRestartDelay", "5000"},
		{"start", SERVICE_NAME_RPC},
		// winagentsvc
		{"install", SERVICE_NAME_AGENT, a.EXE, "-m", "winagentsvc"},
		{"set", SERVICE_NAME_AGENT, "DisplayName", SERVICE_DESC_AGENT},
		{"set", SERVICE_NAME_AGENT, "Description", SERVICE_DESC_AGENT},
		{"set", SERVICE_NAME_AGENT, "AppRestartDelay", "5000"},
		{"start", SERVICE_NAME_AGENT},
	}

	for _, s := range svcCommands {
		a.Logger.Debugln(a.Nssm, s)
		_, _ = CMD(a.Nssm, s, 25, false)
	}

	// 2022-01-01: optional
	if i.WinDefender {
		a.Logger.Infoln("Adding Windows Defender exclusions")
		a.addDefenderExclusions()
	}

	if i.Power {
		a.Logger.Infoln("Disabling sleep/hibernate...")
		DisableSleepHibernate()
	}

	if i.Ping {
		a.Logger.Infoln("Enabling ping...")
		EnablePing()
	}

	if i.RDP {
		a.Logger.Infoln("Enabling RDP...")
		EnableRDP()
	}

	a.installerMsg("Installation was successfully!\nAllow a few minutes for the agent to show up in the RMM", "info", i.Silent)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) checkExistingAndRemove(silent bool) {
	hasReg := false
	_, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\TacticalRMM`, registry.ALL_ACCESS)
	if err == nil {
		hasReg = true
	}
	installedMesh := filepath.Join(a.ProgramDir, "Mesh Agent", "MeshAgent.exe")
	installedSalt := filepath.Join(a.SystemDrive, "\\salt", "uninst.exe")
	agentDB := filepath.Join(a.ProgramDir, "agentdb.db")
	if hasReg || FileExists(installedMesh) || FileExists(installedSalt) || FileExists(agentDB) {
		tacUninst := filepath.Join(a.ProgramDir, a.GetUninstallExe())
		tacUninstArgs := []string{tacUninst, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/FORCECLOSEAPPLICATIONS"}

		window := w32.GetForegroundWindow()
		if !silent && window != 0 {
			var handle w32.HWND
			msg := "Existing installation found\nClick OK to remove, then re-run the installer.\nClick Cancel to abort."
			action := w32.MessageBox(handle, msg, "Tactical RMM", w32.MB_OKCANCEL|w32.MB_ICONWARNING)
			if action == w32.IDOK {
				a.AgentUninstall()
			}
		} else {
			fmt.Println("Existing installation found and must be removed before attempting to reinstall.")
			fmt.Println("Run the following command to uninstall, and then re-run this installer.")
			fmt.Printf(`"%s" %s %s %s`, tacUninstArgs[0], tacUninstArgs[1], tacUninstArgs[2], tacUninstArgs[3])
		}
		os.Exit(0)
	}
}
