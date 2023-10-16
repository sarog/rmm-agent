package windows

import (
	"fmt"
	wapi "github.com/jetrmm/go-win64api"
	"github.com/sarog/rmmagent/agent/common"
	rmm "github.com/sarog/rmmagent/shared"
)

func (a *windowsAgent) GetInstalledSoftware() []rmm.SoftwareList {
	ret := make([]rmm.SoftwareList, 0)

	sw, err := wapi.InstalledSoftwareList()
	if err != nil {
		return ret
	}

	for _, s := range sw {
		t := s.InstallDate
		ret = append(ret, rmm.SoftwareList{
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
