package agent

import (
	"github.com/nats-io/nats.go"
	"time"
)

// SetupNatsOptions
// todo: review authentication
// https://docs.nats.io/running-a-nats-service/configuration/securing_nats/auth_intro/nkey_auth
func (a *Agent) SetupNatsOptions() []nats.Option {
	totalWait := 10 * time.Minute
	opts := make([]nats.Option, 0)
	opts = append(opts, nats.Name(a.AgentID))
	opts = append(opts, nats.UserInfo(a.AgentID, a.Token))
	opts = append(opts, nats.ReconnectWait(time.Second*5))
	opts = append(opts, nats.RetryOnFailedConnect(true))
	opts = append(opts, nats.MaxReconnects(-1))
	opts = append(opts, nats.ReconnectBufSize(-1))
	// opts = append(opts, nats.PingInterval(time.Duration(a.NatsPingInterval)*time.Second))
	// opts = append(opts, nats.Compression(a.NatsWSCompression))
	// opts = append(opts, nats.ProxyPath(a.NatsProxyPath))
	// opts = append(opts, nats.ReconnectJitter(500*time.Millisecond, 4*time.Second))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		a.Logger.Printf("NATS Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		a.Logger.Printf("NATS Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		a.Logger.Fatalf("NATS Error: %v", err)
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		a.Logger.Fatalf("NATS Exiting: %v", nc.LastError())
	}))
	// if a.Insecure {
	// 	insecureConf := &tls.Config{
	// 		InsecureSkipVerify: true,
	// 	}
	// 	opts = append(opts, nats.Secure(insecureConf))
	// }
	return opts
}
