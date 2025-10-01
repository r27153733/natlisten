package natnet

import (
	"context"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/r27153733/natlisten/natnet/internal/metricset"
)

type KeepAliveConf struct {
	Timeout            time.Duration
	HttpKeepAliveSleep time.Duration
	LocalAddr          net.Addr
	HTTPAddrPort       []string
	TcpKeepAliveConf   TcpKeepAliveConf
	Network            string
	IgnoreErrCnt       uint32
}

func initKeepAliveConf(c *KeepAliveConf) {
	if c.Timeout == 0 {
		c.Timeout = time.Second * 3
	}
	if c.HttpKeepAliveSleep == 0 {
		c.HttpKeepAliveSleep = time.Second * 15
	}
	if c.LocalAddr == nil {
		c.LocalAddr = &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 26656,
		}
	}
	if len(c.HTTPAddrPort) == 0 {
		c.HTTPAddrPort = []string{
			"qq.com:80",
			"baidu.com:80",
		}
	}
	if c.Network == "" {
		c.Network = "tcp"
	}
	if c.IgnoreErrCnt == 0 {
		c.IgnoreErrCnt = 3
	}
}

func StartHttpKeepAlive(ctx context.Context, conf KeepAliveConf) error {
	metricset.KeepaliveGoroutine.Inc()

	buf := make([]byte, 0, 96)
	initKeepAliveConf(&conf)
	t := time.NewTicker(conf.HttpKeepAliveSleep)
	defer t.Stop()
	var serverIdx uint
	for {
		select {
		case <-ctx.Done():
			metricset.KeepaliveGoroutine.Dec()
			return ctx.Err()
		default:
			_ = startHttpKeepAlive(ctx, conf, t, serverIdx, buf)
			serverIdx++
			metricset.KeepaliveSwitch.Inc()
		}
	}
}

func startHttpKeepAlive(ctx context.Context, conf KeepAliveConf, t *time.Ticker, serverIdx uint, buf []byte) error {
	conn, err := getKeepAliveConn(ctx, conf, serverIdx)
	if err != nil {
		metricset.KeepaliveConnErr.Inc()
		return err
	}
	defer func(cc net.Conn) {
		err := cc.Close()
		if err != nil {
			metricset.KeepaliveConnErr.Inc()
		}
	}(conn)

	buf = append(buf[:0], "HEAD / HTTP/1.1\r\nHost: "...)
	buf = append(buf, conf.HTTPAddrPort[serverIdx%uint(len(conf.HTTPAddrPort))]...)
	buf = append(buf, "\r\nConnection: keep-alive\r\n\r\n"...)

	var cnt atomic.Uint32
	cnt.Store(conf.IgnoreErrCnt - 1)

	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		metricset.KeepaliveConnErr.Inc()
	}

	go func() {
		b := make([]byte, 1024)
		for cnt.Load() <= conf.IgnoreErrCnt {
			n, err := conn.Read(b)
			metricset.KeepaliveReadLastTime.Set(float64(time.Now().Unix()))
			if err != nil && err != io.EOF {
				if errOp := (&net.OpError{}); errors.As(err, &errOp) {
					// This means we just closed the connection and it's OK
					if errOp.Err.Error() == "use of closed network connection" {
						return
					}
				}
				metricset.KeepaliveReadErr.Inc()
			}
			metricset.KeepaliveReadBytes.Add(n)

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	for cnt.Load() <= conf.IgnoreErrCnt {
		err = conn.SetWriteDeadline(time.Now().Add(conf.Timeout))
		if err != nil {
			cnt.Add(1)
			metricset.KeepaliveWriteErr.Inc()
		} else {
			n, err := conn.Write(buf)
			metricset.KeepaliveWriteLastTime.Set(float64(time.Now().Unix()))
			if err != nil {
				cnt.Add(1)
			} else {
				cnt.Store(0)
			}
			metricset.KeepaliveWriteBytes.Add(n)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
	return err
}

func getKeepAliveConn(ctx context.Context, conf KeepAliveConf, serverIdx uint) (net.Conn, error) {
	dCtx, dCancelFunc := context.WithDeadline(ctx, time.Now().Add(conf.Timeout))
	defer dCancelFunc()

	conn, err := dialWithReuseKeepAlive(dCtx, conf.LocalAddr, conf.Network, conf.HTTPAddrPort[serverIdx%uint(len(conf.HTTPAddrPort))], &conf.TcpKeepAliveConf)
	if err != nil {
		return nil, err
	}
	return conn, err
}
