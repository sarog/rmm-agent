package agent

import (
	"fmt"
	"time"

	rmm "github.com/sarog/rmmagent/shared"
)

const (
	API_URL_WINUPDATES = "/api/v3/winupdates/"
	API_URL_SUPERSEDED = "/api/v3/superseded/"
)

func (a *Agent) GetWinUpdates() {
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
	// 2022-01-01: api/tacticalrmm/apiv3/views.py:172
	_, err = a.rClient.R().SetBody(payload).Post(API_URL_WINUPDATES)
	if err != nil {
		a.Logger.Debugln(err)
	}
}

func (a *Agent) InstallUpdates(guids []string) {
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
			// 2022-01-01: api/tacticalrmm/apiv3/views.py:148
			a.rClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
			continue
		}
		defer updts.Release()

		updtCnt, err := updts.Count()
		if err != nil {
			a.Logger.Errorln(err)
			result.Success = false
			// 2022-01-01: api/tacticalrmm/apiv3/views.py:148
			a.rClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
			continue
		}
		a.Logger.Debugln("updtCnt:", updtCnt)

		if updtCnt == 0 {
			superseded := rmm.SupersededUpdate{AgentID: a.AgentID, UpdateID: id}
			// 2022-01-01: api/tacticalrmm/apiv3/views.py:220
			a.rClient.R().SetBody(superseded).Post(API_URL_SUPERSEDED)
			continue
		}

		for i := 0; i < int(updtCnt); i++ {
			u, err := updts.Item(i)
			if err != nil {
				a.Logger.Errorln(err)
				result.Success = false
				// 2022-01-01: api/tacticalrmm/apiv3/views.py:148
				a.rClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
				continue
			}
			a.Logger.Debugln("u:", u)
			err = session.InstallWUAUpdate(u)
			if err != nil {
				a.Logger.Errorln(err)
				result.Success = false
				// 2022-01-01: api/tacticalrmm/apiv3/views.py:148
				a.rClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
				continue
			}
			result.Success = true
			// 2022-01-01: api/tacticalrmm/apiv3/views.py:148
			a.rClient.R().SetBody(result).Patch(API_URL_WINUPDATES)
			a.Logger.Debugln("Installed windows update with guid", id)
		}
	}

	time.Sleep(5 * time.Second)
	needsReboot, err := a.SystemRebootRequired()
	if err != nil {
		a.Logger.Errorln(err)
	}

	rebootPayload := rmm.AgentNeedsReboot{AgentID: a.AgentID, NeedsReboot: needsReboot}
	// 2021-12-31: api/tacticalrmm/apiv3/views.py:122
	_, err = a.rClient.R().SetBody(rebootPayload).Put(API_URL_WINUPDATES)
	if err != nil {
		a.Logger.Debugln("NeedsReboot:", err)
	}
}
