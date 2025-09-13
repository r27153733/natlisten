package natnet

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/pion/stun"
)

type STUNProbeConf struct {
	Timeout      time.Duration
	ProbeSleep   time.Duration
	LocalAddr    net.Addr
	STUNAddrPort []string
	Network      string
	IgnoreErrCnt uint32
}

func initSTUNProbeConf(c *STUNProbeConf) {
	if c.Timeout == 0 {
		c.Timeout = 3 * time.Second
	}
	if c.ProbeSleep == 0 {
		c.ProbeSleep = 15 * time.Second
	}
	if c.LocalAddr == nil {
		c.LocalAddr = &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 26656,
		}
	}
	if len(c.STUNAddrPort) == 0 {
		c.STUNAddrPort = []string{
			"turn.cloudflare.com:3478",
			"fwa.lifesizecloud.com:3478",
			"global.turn.twilio.com:3478",
			"stun.isp.net.au:3478",
			"stun.nextcloud.com:3478",
			"stun.freeswitch.org:3478",
			"stun.voip.blackberry.com:3478",
			"stunserver.stunprotocol.org:3478",
			"stun.sipnet.com:3478",
			"stun.radiojar.com:3478",
			"stun.sonetel.com:3478",
			"stun.telnyx.com:3478",
		}
	}
	if c.Network == "" {
		c.Network = "tcp"
	}
	if c.IgnoreErrCnt == 0 {
		c.IgnoreErrCnt = 3
	}
}

func StartSTUNProbe(ctx context.Context, conf STUNProbeConf, fn func(ip net.IP, port int) error) error {
	initSTUNProbeConf(&conf)
	t := time.NewTicker(conf.ProbeSleep)
	defer t.Stop()
	var serverIdx uint
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_ = startSTUNProbe(ctx, conf, t, serverIdx, fn)
			serverIdx++
		}
	}
}

func startSTUNProbe(ctx context.Context, conf STUNProbeConf, t *time.Ticker, serverIdx uint, fn func(ip net.IP, port int) error) error {
	conn, err := getSTUNConn(ctx, conf, serverIdx)
	if err != nil {
		return err
	}

	c, err := stun.NewClient(conn)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var cnt atomic.Uint32
	cnt.Store(conf.IgnoreErrCnt - 1)
	for cnt.Load() <= conf.IgnoreErrCnt {
		message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
		err = c.Start(message, func(res stun.Event) {
			if res.Error != nil {
				cnt.Add(1)
				return
			}

			var xorAddr stun.XORMappedAddress
			err := xorAddr.GetFrom(res.Message)
			if err != nil {
				cnt.Add(1)
				return
			}

			cnt.Store(0)
			println(xorAddr.String())
			_ = fn(xorAddr.IP, xorAddr.Port)
		})
		if err != nil {
			cnt.Add(1)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}

	return errors.New("STUN probe failed")
}

func getSTUNConn(ctx context.Context, conf STUNProbeConf, serverIdx uint) (net.Conn, error) {
	dCtx, dCancelFunc := context.WithDeadline(ctx, time.Now().Add(conf.Timeout))
	defer dCancelFunc()
	conn, err := dialWithReuse(dCtx, conf.LocalAddr, conf.Network, conf.STUNAddrPort[serverIdx%uint(len(conf.STUNAddrPort))])
	if err != nil {
		return nil, err
	}
	return conn, err
}
