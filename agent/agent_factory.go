package agent

import (
	"github.com/sarog/rmmagent/agent/windows"
	"github.com/sirupsen/logrus"
	"runtime"
)

func getAgent(logger *logrus.Logger, version string) IAgent {
	switch runtime.GOOS {
	case "windows":
		return windows.NewAgent(logger, version)
	case "freebsd":
	case "darwin":
	case "linux":
		// todo
	}
}
