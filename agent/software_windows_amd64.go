package agent

import (
	"strings"

	wapi "github.com/iamacarpet/go-win64api"
)

type SoftwareList struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Publisher   string `json:"publisher"`
	InstallDate string `json:"install_date"`
	Size        string `json:"size"`
	Source      string `json:"source"`
	Location    string `json:"location"`
	Uninstall   string `json:"uninstall"`
}

// GetProgramVersion loops through the registry for software
// and if found, returns its version
func (a *WindowsAgent) GetProgramVersion(name string) string {
	sw, err := wapi.InstalledSoftwareList()
	if err != nil {
		a.Logger.Debugf("%s\r\n", err.Error())
		return "0.0.0"
	}

	var lowerName string
	for _, s := range sw {
		lowerName = strings.ToLower(s.Name())
		if strings.Contains(lowerName, name) {
			return s.Version()
		}
	}
	return "0.0.0"
}

func (a *WindowsAgent) GetInstalledSoftware() []SoftwareList {
	ret := make([]SoftwareList, 0)

	sw, err := wapi.InstalledSoftwareList()
	if err != nil {
		return ret
	}

	for _, s := range sw {
		ret = append(ret, SoftwareList{
			Name:        s.Name(),
			Version:     s.Version(),
			Publisher:   s.Publisher,
			InstallDate: s.InstallDate.String(),
			Size:        ByteCountSI(s.EstimatedSize * 1024),
			Source:      s.InstallSource,
			Location:    s.InstallLocation,
			Uninstall:   s.UninstallString,
		})
	}
	return ret
}
