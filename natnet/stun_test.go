package natnet

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	go func() {
		StartSTUNProbe(context.Background(), STUNProbeConf{
			Timeout:      3 * time.Second,
			ProbeSleep:   15 * time.Second,
			LocalAddr:    &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 8880},
			STUNAddrPort: []string{"turn.cloudflare.com:3478"},
		}, func(ip net.IP, port int) {

		})
	}()
	StartHttpKeepAlive(context.Background(), KeepAliveConf{
		Timeout:            3 * time.Second,
		HttpKeepAliveSleep: 15 * time.Second,
		LocalAddr:          &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 8880},
		HTTPAddrPort:       []string{"qq.com:80"},
		TcpKeepAliveConf:   TcpKeepAliveConf{},
	})
}
