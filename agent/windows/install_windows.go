package windows

import (
	"fmt"
	"github.com/sarog/rmmagent/agent/common"
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
	ServerURL   string
	ClientID    int
	SiteID      int
	Description string
	Token       string
	Cert        string
	Timeout     time.Duration
	ApiURL      string // was "SaltMaster"
	Silent      bool

	// Deprecated
	// AgentType string
	// Deprecated
	DisableSleep bool
	// Deprecated
	EnableRDP bool
	// Deprecated
	EnablePing bool
}

func createRegKeys(baseurl, agentid, apiurl, token, agentpk, cert string) {
	// todo: 2021-12-31: migrate to DPAPI
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, REG_RMM_PATH, registry.ALL_ACCESS)
	if err != nil {
		log.Fatalln("Error creating registry key:", err)
	}
	defer key.Close()

	err = key.SetStringValue(REG_RMM_BASEURL, baseurl)
	if err != nil {
		log.Fatalln("Error creating BaseURL registry key:", err)
	}

	err = key.SetStringValue(REG_RMM_AGENTID, agentid)
	if err != nil {
		log.Fatalln("Error creating AgentID registry key:", err)
	}

	err = key.SetStringValue(REG_RMM_APIURL, apiurl)
	if err != nil {
		log.Fatalln("Error creating ApiURL registry key:", err)
	}

	err = key.SetStringValue(REG_RMM_TOKEN, token)
	if err != nil {
		log.Fatalln("Error creating Token registry key:", err)
	}

	err = key.SetStringValue(REG_RMM_AGENTPK, agentpk)
	if err != nil {
		log.Fatalln("Error creating AgentPK registry key:", err)
	}

	if len(cert) > 0 {
		err = key.SetStringValue(REG_RMM_CERT, cert)
		if err != nil {
			log.Fatalln("Error creating RootCert registry key:", err)
		}
	}
}

func (a *windowsAgent) Install(i *common.InstallInfo, agentID string) {
	a.checkExistingAndRemove(i.Silent)

	i.Headers = map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Token %s", i.Token),
	}
	a.AgentID = agentID
	a.Logger.Debugln("Agent ID:", a.AgentID)

	u, err := url.Parse(i.ServerURL)
	if err != nil {
		a.installerMsg(err.Error(), "error", i.Silent)
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		a.installerMsg("Invalid URL: must begin with https or http", "error", i.Silent)
	}

	// This will match either IPv4 or IPv4:port
	var ipPort = regexp.MustCompile(`[0-9]+(?:\.[0-9]+){3}(:[0-9]+)?`)

	// if ipv4:port, strip the port to get ip for salt master
	if ipPort.MatchString(u.Host) && strings.Contains(u.Host, ":") {
		i.ApiURL = strings.Split(u.Host, ":")[0]
	} else if strings.Contains(u.Host, ":") {
		i.ApiURL = strings.Split(u.Host, ":")[0]
	} else {
		i.ApiURL = u.Host
	}

	a.Logger.Debugln("API URL:", i.ApiURL)

	terr := common.TestTCP(fmt.Sprintf("%s:4222", i.ApiURL))
	if terr != nil {
		a.installerMsg(fmt.Sprintf("ERROR: Either port 4222 TCP is not open on your RMM server, or nats.service is not running.\n\n%s", terr.Error()), "error", i.Silent)
	}

	baseURL := u.Scheme + "://" + u.Host
	a.Logger.Debugln("Base URL:", baseURL)

	iClient := resty.New()
	iClient.SetCloseConnection(true)
	iClient.SetTimeout(15 * time.Second)
	iClient.SetDebug(a.Debug)
	iClient.SetHeaders(i.Headers)
	creds, cerr := iClient.R().Get(fmt.Sprintf("%s/api/v3/installer/", baseURL))
	if cerr != nil {
		a.installerMsg(cerr.Error(), "error", i.Silent)
	}
	if creds.StatusCode() == 401 {
		a.installerMsg("Installer token has expired. Please generate a new one.", "error", i.Silent)
	}

	verPayload := map[string]string{"version": a.Version}

	iVersion, ierr := iClient.R().SetBody(verPayload).Post(fmt.Sprintf("%s/api/v3/installer/", baseURL))
	if ierr != nil {
		a.installerMsg(ierr.Error(), "error", i.Silent)
	}
	if iVersion.StatusCode() != 200 {
		a.installerMsg(common.DjangoStringResp(iVersion.String()), "error", i.Silent)
	}

	rClient := resty.New()
	rClient.SetCloseConnection(true)
	rClient.SetTimeout(i.Timeout * time.Second)
	rClient.SetDebug(a.Debug)
	// Set REST knox headers
	rClient.SetHeaders(i.Headers)

	// Set local certificate if applicable
	if len(i.RootCert) > 0 {
		if !common.FileExists(i.RootCert) {
			a.installerMsg(fmt.Sprintf("%s does not exist", i.RootCert), "error", i.Silent)
		}
		rClient.SetRootCertificate(i.RootCert)
	}

	a.Logger.Infoln("Adding agent to the dashboard")

	type NewAgentResp struct {
		AgentPK int `json:"pk"`
		// SaltID  string `json:"saltid"`
		Token string `json:"token"`
	}

	agentPayload := map[string]interface{}{
		"agent_id":    a.AgentID,
		"hostname":    a.Hostname(),
		"client":      i.ClientID,
		"site":        i.SiteID,
		"description": i.Description,
		// -sg: "monitoring_type": i.AgentType,
	}

	r, err := rClient.R().SetBody(agentPayload).SetResult(&NewAgentResp{}).Post(fmt.Sprintf("%s/api/v3/newagent/", baseURL))
	if err != nil {
		a.installerMsg(err.Error(), "error", i.Silent)
	}
	if r.StatusCode() != 200 {
		a.installerMsg(r.String(), "error", i.Silent)
	}

	agentPK := r.Result().(*NewAgentResp).AgentPK
	agentToken := r.Result().(*NewAgentResp).Token

	a.Logger.Debugln("Agent Token:", agentToken)
	a.Logger.Debugln("Agent PK:", agentPK)

	// todo: extract / move
	createRegKeys(baseURL, a.AgentID, i.ApiURL, agentToken, strconv.Itoa(agentPK), i.RootCert)

	// Refresh our agent with new values
	a = a.New(a.Logger, a.Version)
	// todo:
	// a = agent.GetAgent(a.Logger, a.Version)
	// a = NewAgent(a.Logger, a.Version)

	// Set new headers. No longer knox auth; use agent auth
	rClient.SetHeaders(a.Headers)

	// Send WMI system information
	a.Logger.Debugln("Getting system information with WMI")
	a.SysInfo()

	// Check in once via nats
	opts := a.SetupNatsOptions()
	server := fmt.Sprintf("tls://%s:%d", a.ApiURL, a.ApiPort)

	nc, err := nats.Connect(server, opts...)
	if err != nil {
		a.Logger.Errorln(err)
	} else {
		startup := []string{CHECKIN_MODE_HELLO, CHECKIN_MODE_OSINFO, CHECKIN_MODE_WINSERVICES, CHECKIN_MODE_DISKS, CHECKIN_MODE_PUBLICIP, CHECKIN_MODE_SOFTWARE, CHECKIN_MODE_LOGGEDONUSER}
		for _, mode := range startup {
			a.CheckIn(nc, mode)
			time.Sleep(200 * time.Millisecond)
		}
		nc.Close()
	}

	a.Logger.Debugln("Creating temporary directory")
	a.CreateAgentTempDir()

	a.Logger.Infoln("Installing services...")

	// todo
	/*svcCommands := [10][]string{
		// rpc // todo: remove/combine
		{"install", SERVICE_NAME_RPC, a.AgentExe, "-m", AGENT_MODE_RPC},
		{"set", SERVICE_NAME_RPC, "DisplayName", SERVICE_DESC_RPC},
		{"set", SERVICE_NAME_RPC, "Description", SERVICE_DESC_RPC},
		{"set", SERVICE_NAME_RPC, "AppRestartDelay", SERVICE_RESTART_DELAY},
		{"start", SERVICE_NAME_RPC},
		// winagentsvc
		{"install", SERVICE_NAME_AGENT, a.AgentExe, "-m", AGENT_MODE_SVC},
		{"set", SERVICE_NAME_AGENT, "DisplayName", SERVICE_DESC_AGENT},
		{"set", SERVICE_NAME_AGENT, "Description", SERVICE_DESC_AGENT},
		{"set", SERVICE_NAME_AGENT, "AppRestartDelay", SERVICE_RESTART_DELAY},
		{"start", SERVICE_NAME_AGENT},
	}

	for _, s := range svcCommands {
		a.Logger.Debugln(a.Nssm, s)
		_, _ = CMD(a.Nssm, s, 25, false)
	}*/

	a.installerMsg("Installation was successful!\nPlease allow a few minutes for the agent to show up in the RMM server", "info", i.Silent)
}

// todo: add to Agent interface
func (a *windowsAgent) checkExistingAndRemove(silent bool) {
	hasReg := false
	_, err := registry.OpenKey(registry.LOCAL_MACHINE, REG_RMM_PATH, registry.ALL_ACCESS)
	if err == nil {
		hasReg = true
	}
	if hasReg {
		tacUninst := filepath.Join(a.ProgramDir, a.GetUninstallExe())
		tacUninstArgs := []string{tacUninst, "/VERYSILENT", "/SUPPRESSMSGBOXES", "/FORCECLOSEAPPLICATIONS"}

		window := w32.GetForegroundWindow()
		if !silent && window != 0 {
			var handle w32.HWND
			msg := "Existing installation found\nClick OK to remove, then re-run the installer.\nClick Cancel to abort."
			action := w32.MessageBox(handle, msg, common.AGENT_NAME_LONG, w32.MB_OKCANCEL|w32.MB_ICONWARNING)
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
