package windows

const API_URL_SYSINFO = "/api/v3/sysinfo/"

// SysInfo Retrieves (and sends) system information
func (a *windowsAgent) SysInfo() {
	wmiInfo := make(map[string]interface{})

	compSysProd, err := GetWin32_ComputerSystemProduct()
	if err != nil {
		a.Logger.Debugln(err)
	}

	compSys, err := GetWin32_ComputerSystem()
	if err != nil {
		a.Logger.Debugln(err)
	}

	netAdaptConfig, err := GetWin32_NetworkAdapterConfiguration()
	if err != nil {
		a.Logger.Debugln(err)
	}

	physMem, err := GetWin32_PhysicalMemory()
	if err != nil {
		a.Logger.Debugln(err)
	}

	winOS, err := GetWin32_OperatingSystem()
	if err != nil {
		a.Logger.Debugln(err)
	}

	baseBoard, err := GetWin32_BaseBoard()
	if err != nil {
		a.Logger.Debugln(err)
	}

	bios, err := GetWin32_BIOS()
	if err != nil {
		a.Logger.Debugln(err)
	}

	disk, err := GetWin32_DiskDrive()
	if err != nil {
		a.Logger.Debugln(err)
	}

	netAdapt, err := GetWin32_NetworkAdapter()
	if err != nil {
		a.Logger.Debugln(err)
	}

	desktopMon, err := GetWin32_DesktopMonitor()
	if err != nil {
		a.Logger.Debugln(err)
	}

	cpu, err := GetWin32_Processor()
	if err != nil {
		a.Logger.Debugln(err)
	}

	usb, err := GetWin32_USBController()
	if err != nil {
		a.Logger.Debugln(err)
	}

	graphics, err := GetWin32_VideoController()
	if err != nil {
		a.Logger.Debugln(err)
	}

	wmiInfo["comp_sys_prod"] = compSysProd
	wmiInfo["comp_sys"] = compSys
	wmiInfo["network_config"] = netAdaptConfig
	wmiInfo["mem"] = physMem
	wmiInfo["os"] = winOS
	wmiInfo["base_board"] = baseBoard
	wmiInfo["bios"] = bios
	wmiInfo["disk"] = disk
	wmiInfo["network_adapter"] = netAdapt
	wmiInfo["desktop_monitor"] = desktopMon
	wmiInfo["cpu"] = cpu
	wmiInfo["usb"] = usb
	wmiInfo["graphics"] = graphics

	payload := map[string]interface{}{
		"agent_id": a.AgentID,
		"sysinfo":  wmiInfo,
	}

	_, rerr := a.RClient.R().SetBody(payload).Patch(API_URL_SYSINFO)
	if rerr != nil {
		a.Logger.Debugln(rerr)
	}
}
