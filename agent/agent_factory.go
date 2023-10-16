package agent

import (
	"github.com/sarog/rmmagent/agent/common"
	"github.com/sarog/rmmagent/agent/windows"
	"github.com/sirupsen/logrus"
	"runtime"
)

func GetAgent(logger *logrus.Logger, version string) common.IAgent {
	switch runtime.GOOS {
	case "windows":
		return windows.NewAgent(logger, version)
	case "freebsd":
	case "darwin":
	case "linux":
		// todo
	}
	return nil
}
