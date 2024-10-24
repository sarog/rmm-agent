package agent

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

var (
	agentProvider AgentProvider
)

type AgentProvider interface {
	Agent(logger *logrus.Logger, version string) *IAgent
}

func Register(provider interface{}) {
	if a, ok := provider.(AgentProvider); ok {
		if agentProvider != nil {
			panic(fmt.Sprintf("AgentProvider already registered: %v", agentProvider))
		}
		agentProvider = a
	}
}

func GetAgentProvider() AgentProvider { return agentProvider }
