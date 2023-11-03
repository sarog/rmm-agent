package windows

import (
	"time"

	"github.com/go-resty/resty/v2"

	rmm "github.com/sarog/rmmagent/shared"
)

const API_URL_CHOCO = "/api/v3/choco/"

func (a *windowsAgent) InstallPkgMgr(pkgMgr string) {
	// todo: choco, scoop, winget, custom?

	switch pkgMgr {
	case "choco":
		a.installChoco()
	}
}

func (a *windowsAgent) RemovePkgMgr(pkgMgr string) {

}

func (a *windowsAgent) InstallPackage(pkgMgr string, pkgName string) (string, error) {
	switch pkgMgr {
	case "choco":
		return a.installWithChoco(pkgName)
	case "scoop":
	case "winget":
	}

	return "", nil
}

func (a *windowsAgent) RemovePackage(pkgMgr string, pkgName string) (string, error) {
	switch pkgMgr {
	case "choco":
		return a.installWithChoco(pkgName)
	case "scoop":
	case "winget":
	}

	return "", nil
}

func (a *windowsAgent) UpdatePackage(pkgMgr string, pkgName string) (string, error) {
	switch pkgMgr {
	case "choco":
		return a.installWithChoco(pkgName)
	case "scoop":
	case "winget":
	}

	return "", nil
}

// installChoco Installs the Chocolatey Package Manager using PowerShell
func (a *windowsAgent) installChoco() {
	// todo: see: https://docs.chocolatey.org/en-us/choco/setup

	var result rmm.PkgMgrInstalled
	result.AgentID = a.AgentID
	result.Installed = false
	result.PackageManager = "choco"

	rClient := resty.New()
	rClient.SetTimeout(30 * time.Second)

	r, err := rClient.R().Get("https://chocolatey.org/install.ps1")
	if err != nil {
		a.Logger.Debugln(err)
		a.RClient.R().SetBody(result).Post(API_URL_CHOCO)
		return
	}
	if r.IsError() {
		a.RClient.R().SetBody(result).Post(API_URL_CHOCO)
		return
	}

	_, _, exitcode, err := a.RunScript(string(r.Body()), "powershell", []string{}, 900)
	if err != nil {
		a.Logger.Debugln(err)
		a.RClient.R().SetBody(result).Post(API_URL_CHOCO)
		return
	}

	if exitcode != 0 {
		a.RClient.R().SetBody(result).Post(API_URL_CHOCO)
		return
	}

	result.Installed = true
	a.RClient.R().SetBody(result).Post(API_URL_CHOCO)
}

// installWithChoco install an application with Chocolatey
func (a *windowsAgent) installWithChoco(name string) (string, error) {
	out, err := CMD("choco.exe", []string{"install", name, "--yes", "--force", "--force-dependencies"}, 1200, false)
	if err != nil {
		a.Logger.Errorln(err)
		return err.Error(), err
	}
	if out[1] != "" {
		return out[1], nil
	}
	return out[0], nil
}

// todo
func (a *windowsAgent) removeWithChoco(name string) (string, error) {
	out, err := CMD("choco.exe", []string{"install", name, "--yes", "--force", "--force-dependencies"}, 1200, false)
	if err != nil {
		a.Logger.Errorln(err)
		return err.Error(), err
	}
	if out[1] != "" {
		return out[1], nil
	}
	return out[0], nil
}

// todo
func (a *windowsAgent) updateWithChoco(name string) (string, error) {
	out, err := CMD("choco.exe", []string{"install", name, "--yes", "--force", "--force-dependencies"}, 1200, false)
	if err != nil {
		a.Logger.Errorln(err)
		return err.Error(), err
	}
	if out[1] != "" {
		return out[1], nil
	}
	return out[0], nil
}
