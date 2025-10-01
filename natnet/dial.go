package natnet

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func DialWithReuse(ctx context.Context, localAddr net.Addr, network, remote string) (net.Conn, error) {
	d := net.Dialer{
		LocalAddr:      localAddr,
		ControlContext: controlReuse,
	}

	return d.DialContext(ctx, network, remote)
}

func controlReuse(ctx context.Context, network, address string, c syscall.RawConn) error {
	return c.Control(setReuse)
}

func setReuse(fd uintptr) {
	_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}

type TcpKeepAliveConf struct {
	IdleSec, IntervalSec, Count int
}

func dialWithReuseKeepAlive(ctx context.Context, local net.Addr, network, remote string, conf *TcpKeepAliveConf) (net.Conn, error) {
	d := net.Dialer{
		LocalAddr: local,
		ControlContext: func(ctx context.Context, network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				setReuse(fd)
				_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)

				if conf.IdleSec != 0 {
					_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, conf.IdleSec)
				}

				if conf.IntervalSec != 0 {
					_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, conf.IntervalSec)
				}

				if conf.Count != 0 {
					_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPCNT, conf.Count)
				}
			})
		},
	}

	return d.DialContext(ctx, network, remote)
}
