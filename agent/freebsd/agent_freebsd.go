package freebsd

type freebsdAgent struct {
	common.Agent
}

func NewAgent(logger *logrus.Logger, version string) common.IAgent {

}

func (a *freebsdAgent) Install(i *common.InstallInfo, agentID string) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) InstallService() error {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) AgentUpdate(url, inno, version string) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) AgentUninstall() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) UninstallCleanup() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RunAgentService(nc *nats.Conn) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RunService() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) ShowStatus(version string) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RunTask(i int) error {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RunChecks(force bool) error {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RunScript(code string, shell string, args []string, timeout int) (stdout, stderr string, exitcode int, e error) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) CheckIn(nc *nats.Conn, mode string) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) CreateInternalTask(name, args, repeat string, start int) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) CheckRunner() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) GetCheckInterval() (int, error) {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) SendSoftware() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) SyncInfo() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RecoverAgent() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) GetServiceConfig() *service.Config {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) PublicIP() string {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) TotalRAM() float64 {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) BootTime() int64 {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) GetInstalledSoftware() []jrmm.Software {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) OSInfo() (plat, osFullName string) {
	// TODO implement me
	// uname -a
	panic("implement me")
}

func (a *freebsdAgent) SysInfo() {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) GetStorage() []jrmm.StorageDrive {
	// TODO implement me
	// df
	panic("implement me")
}

func (a *freebsdAgent) LoggedOnUser() string {
	// TODO implement me
	// whoami
	panic("implement me")
}

func (a *freebsdAgent) GetCPULoadAvg() int {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) Start(s service.Service) error {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) Stop(s service.Service) error {
	// TODO implement me
	panic("implement me")
}

func (a *freebsdAgent) RebootSystem() {
	// TODO implement me
	// shutdown -r now
	panic("implement me")
}
