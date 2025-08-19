package natnet

import (
	"context"
	"fmt"
	"net"
)

func IPV4PubNat(ctx context.Context, localAddr *net.TCPAddr, fn func(ip net.IP, port int)) error {
	ipv4LocalAddr := &net.TCPAddr{
		IP:   localAddr.IP.To4(),
		Port: localAddr.Port,
	}

	if ipv4LocalAddr.IP == nil {
		for i := range localAddr.IP {
			if localAddr.IP[i] != 0 {
				return fmt.Errorf("invalid IPv4 address: %s", ipv4LocalAddr.String())
			}
		}
		ipv4LocalAddr.IP = make(net.IP, net.IPv4len)
	}

	go func() {
		_ = StartSTUNProbe(ctx, STUNProbeConf{
			LocalAddr: ipv4LocalAddr,
			Network:   "tcp4",
		}, fn)
	}()

	go func() {
		_ = StartHttpKeepAlive(ctx, KeepAliveConf{
			LocalAddr: ipv4LocalAddr,
			Network:   "tcp4",
		})
	}()

	return nil
}
