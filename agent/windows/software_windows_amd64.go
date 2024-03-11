package windows

import (
	"fmt"
	wapi "github.com/jetrmm/go-win64api"
	"github.com/jetrmm/rmm-agent/agent/common"
	rmm "github.com/jetrmm/rmm-shared"
)

func (a *windowsAgent) GetInstalledSoftware() []rmm.Software {
	ret := make([]rmm.Software, 0)

	sw, err := wapi.InstalledSoftwareList()
	if err != nil {
		return ret
	}

	for _, s := range sw {
		t := s.InstallDate
		ret = append(ret, rmm.Software{
			Name:        s.Name(),
			Version:     s.Version(),
			Publisher:   s.Publisher,
			InstallDate: fmt.Sprintf("%02d-%d-%02d", t.Year(), t.Month(), t.Day()),
			Size:        common.ByteCountSI(s.EstimatedSize * 1024),
			Source:      s.InstallSource,
			Location:    s.InstallLocation,
			Uninstall:   s.UninstallString,
		})
	}
	return ret
}
