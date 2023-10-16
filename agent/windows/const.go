package windows

// todo: 2021-12-31: custom branding
const (
	SERVICE_NAME_RPC = "rpcagent"
	SERVICE_DESC_RPC = "RMM RPC Service"

	SERVICE_NAME_AGENT = "jetagent"
	SERVICE_DISP_AGENT = "JetRMM Agent Service"
	SERVICE_DESC_AGENT = "JetRMM Agent Service"

	SERVICE_RESTART_DELAY = "5s"

	AGENT_MODE_RPC = "rpc"
	AGENT_MODE_SVC = "agentsvc"

	// Registry strings
	REG_RMM_PATH    = `SOFTWARE\RMMAgent`
	REG_RMM_BASEURL = "BaseURL"
	REG_RMM_AGENTID = "AgentID"
	REG_RMM_APIURL  = "ApiURL"
	REG_RMM_TOKEN   = "Token"
	REG_RMM_AGENTPK = "AgentPK"
	REG_RMM_CERT    = "RootCert"

	AGENT_FOLDER      = "RMMAgent"
	RMM_SEARCH_PREFIX = "acmermm*"
)
