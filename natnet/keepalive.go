package natnet

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"
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
	buf := make([]byte, 0, 96)
	initKeepAliveConf(&conf)
	t := time.NewTicker(conf.HttpKeepAliveSleep)
	defer t.Stop()
	var serverIdx uint
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_ = startHttpKeepAlive(ctx, conf, t, serverIdx, buf)
			serverIdx++
		}
	}
}

func startHttpKeepAlive(ctx context.Context, conf KeepAliveConf, t *time.Ticker, serverIdx uint, buf []byte) error {
	conn, err := getKeepAliveConn(ctx, conf, serverIdx)
	if err != nil {
		return err
	}
	buf = append(buf[:0], "HEAD / HTTP/1.1\r\nHost: "...)
	buf = append(buf, conf.HTTPAddrPort[serverIdx%uint(len(conf.HTTPAddrPort))]...)
	buf = append(buf, "\r\nConnection: keep-alive\r\n\r\n"...)

	var cnt atomic.Uint32
	cnt.Store(conf.IgnoreErrCnt - 1)

	done := ctx.Done()
	go func() {
		for cnt.Load() <= conf.IgnoreErrCnt {
			_, _ = io.Copy(io.Discard, conn)
			select {
			case <-done:
				return
			default:
			}
			time.Sleep(conf.Timeout)
		}
	}()

	for cnt.Load() <= conf.IgnoreErrCnt {
		println("KeepAlive", err, cnt.Load())
		err = conn.SetDeadline(time.Now().Add(conf.Timeout))
		if err != nil {
			cnt.Add(1)
		} else {
			_, err = conn.Write(buf)
			if err != nil {
				cnt.Add(1)
			} else {
				cnt.Store(0)
			}
		}

		select {
		case <-done:
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
