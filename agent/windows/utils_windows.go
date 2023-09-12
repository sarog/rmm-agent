package windows

import (
	"github.com/sarog/rmmagent/agent"
	"math"
	"net"
	"time"

	ps "github.com/elastic/go-sysinfo"
	"github.com/go-resty/resty/v2"
)

// PublicIP returns the agent's public IP address
// Tries 3 times before giving up
func (a *WindowsAgent) PublicIP() string {
	a.Logger.Debugln("PublicIP start")
	client := resty.New()
	client.SetTimeout(4 * time.Second)
	// todo: 2021-12-31: allow custom URLs for IP lookups
	urls := []string{"https://icanhazip.com", "https://ifconfig.co/ip"}
	ip := "error"

	for _, url := range urls {
		r, err := client.R().Get(url)
		if err != nil {
			a.Logger.Debugln("PublicIP error", err)
			continue
		}
		ip = agent.StripAll(r.String())
		if !agent.IsValidIP(ip) {
			a.Logger.Debugln("PublicIP not valid", ip)
			continue
		}
		v4 := net.ParseIP(ip)
		if v4.To4() == nil {
			r1, err := client.R().Get("https://ifconfig.me/ip")
			if err != nil {
				return ip
			}
			ipv4 := agent.StripAll(r1.String())
			if !agent.IsValidIP(ipv4) {
				continue
			}
			a.Logger.Debugln("Forcing IPv4:", ipv4)
			return ipv4
		}
		a.Logger.Debugln("PublicIP return: ", ip)
		break
	}
	return ip
}

// TotalRAM returns total RAM in GB
func (a *WindowsAgent) TotalRAM() float64 {
	host, err := ps.Host()
	if err != nil {
		return 8.0
	}
	mem, err := host.Memory()
	if err != nil {
		return 8.0
	}
	return math.Ceil(float64(mem.Total) / 1073741824.0)
}

// BootTime returns system boot time as a Unix timestamp
func (a *WindowsAgent) BootTime() int64 {
	host, err := ps.Host()
	if err != nil {
		return 1000
	}
	info := host.Info()
	return info.BootTime.Unix()
}
