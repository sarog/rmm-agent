package windows

import (
	"fmt"
	"time"

	rmm "github.com/jetrmm/rmm-agent/shared"
)

const (
	API_URL_WINUPDATES = "/api/v3/winupdates/"
	API_URL_SUPERSEDED = "/api/v3/superseded/"
)

func (a *windowsAgent) GetWinUpdates() {
	updates, err := WUAUpdates("IsInstalled=1 or IsInstalled=0 and Type='Software' and IsHidden=0")
	if err != nil {
		a.Logger.Errorln(err)
		return
	}

	for _, update := range updates {
		a.Logger.Debugln("GUID:", update.UpdateID)
		a.Logger.Debugln("Downloaded:", update.Downloaded)
		a.Logger.Debugln("Installed:", update.Installed)
		a.Logger.Debugln("KB:", update.KBArticleIDs)
		a.Logger.Debugln("--------------------------------")
	}

	payload := rmm.WinUpdateResult{AgentID: a.AgentID, Updates: updates}
	_, err = a.RClient.R().SetBody(payload).Post(API_URL_WINUPDATES)
	if err != nil {
		a.Logger.Debugln(err)
	}
}

func (a *windowsAgent) InstallUpdates(guids []string) {
	session, err := NewUpdateSession()
	if err != nil {
		a.Logger.Errorln(err)
		return
	}
	defer session.Close()

	for _, id := range guids {
		var result rmm.WinUpdateInstallResult
		result.AgentID = a.AgentID
		result.UpdateID = id

		query := fmt.Sprintf("UpdateID='%s'", id)
		a.Logger.Debugln("query:", query)
		updts, err := session.GetWUAUpdateCollection(query)
		if err != nil {
			a.Logger.Errorln(err)
			result.Success = false
			a.RClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
			continue
		}
		defer updts.Release()

		updtCnt, err := updts.Count()
		if err != nil {
			a.Logger.Errorln(err)
			result.Success = false
			a.RClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
			continue
		}
		a.Logger.Debugln("updtCnt:", updtCnt)

		if updtCnt == 0 {
			superseded := rmm.SupersededUpdate{AgentID: a.AgentID, UpdateID: id}
			a.RClient.R().SetBody(superseded).Post(API_URL_SUPERSEDED)
			continue
		}

		for i := 0; i < int(updtCnt); i++ {
			u, err := updts.Item(i)
			if err != nil {
				a.Logger.Errorln(err)
				result.Success = false
				a.RClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
				continue
			}
			a.Logger.Debugln("u:", u)
			err = session.InstallWUAUpdate(u)
			if err != nil {
				a.Logger.Errorln(err)
				result.Success = false
				a.RClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
				continue
			}
			result.Success = true
			a.RClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
			a.Logger.Debugln("Installed Windows update with GUID", id)
		}
	}

	time.Sleep(5 * time.Second)
	needsReboot, err := a.SystemRebootRequired()
	if err != nil {
		a.Logger.Errorln(err)
	}

	rebootPayload := rmm.AgentNeedsReboot{AgentID: a.AgentID, NeedsReboot: needsReboot}
	_, err = a.RClient.R().SetBody(rebootPayload).Put(API_URL_WINUPDATES)
	if err != nil {
		a.Logger.Debugln("NeedsReboot:", err)
	}
}
