module github.com/jetrmm/rmm-agent

go 1.21.3

toolchain go1.22.1

require (
	github.com/go-ole/go-ole v1.3.0
	github.com/go-resty/resty/v2 v2.12.0
	github.com/gonutz/w32/v2 v2.11.1
	github.com/jetrmm/go-dpapi v0.1.0
	github.com/jetrmm/go-sysinfo v0.1.0
	github.com/jetrmm/go-taskmaster v0.1.0
	github.com/jetrmm/go-win64api v0.1.0
	github.com/jetrmm/go-wintoken v0.1.0
	github.com/jetrmm/go-wmi v0.1.0
	github.com/jetrmm/rmm-shared v0.0.0-20231026210319-d19a6850ab00
	github.com/kardianos/service v1.2.2
	github.com/nats-io/nats.go v1.33.1
	github.com/oklog/ulid/v2 v2.1.0
	github.com/shirou/gopsutil/v3 v3.24.2
	github.com/sirupsen/logrus v1.9.3
	github.com/ugorji/go/codec v1.2.12
	golang.org/x/sys v0.19.0
)

require (
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/google/cabbie v1.0.5 // indirect
	github.com/google/glazier v0.0.0-20230912201418-e61e8c721b6f // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20231016141302-07b5767bb0ed // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rickb777/date v1.20.6 // indirect
	github.com/rickb777/plural v1.4.1 // indirect
	github.com/scjalliance/comshim v0.0.0-20230315213746-5e51f40bd3b9 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	gopkg.in/toast.v1 v1.0.0-20180812000517-0a84660828b2 // indirect
	howett.net/plist v1.0.0 // indirect
)

// replace github.com/fourcorelabs/wintoken => github.com/jetrmm/go-wintoken main
// replace github.com/yusufpapurcu/wmi => github.com/jetrmm/go-wmi main
// replace github.com/elastic/go-sysinfo => github.com/jetrmm/go-sysinfo main
