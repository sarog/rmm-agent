package registry

import (
	"fmt"
	"github.com/jetrmm/rmm-agent/agent"
	"github.com/sirupsen/logrus"
)

var (
	agentProvider AgentProvider
)

type AgentProvider interface {
	Agent(logger *logrus.Logger, version string) *agent.IAgent
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
