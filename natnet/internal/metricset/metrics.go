package metricset

import "github.com/VictoriaMetrics/metrics"

var (
	NatSet = metrics.NewSet()

	StunProbeGoroutine        = NatSet.NewCounter(`natnet_stun_probe_goroutine_total`)
	StunProbeSwitch           = NatSet.NewCounter(`natnet_stun_probe_switch_total`)
	StunProbeConnErr          = NatSet.NewCounter(`natnet_stun_probe_err_total{type="conn"}`)
	StunProbeSTUNErr          = NatSet.NewCounter(`natnet_stun_probe_err_total{type="stun"}`)
	StunProbeCallbackErr      = NatSet.NewCounter(`natnet_stun_probe_err_total{type="callback"}`)
	StunProbeCallbackLastTime = NatSet.NewGauge(`natnet_stun_probe_last_time_unix{type="callback",stage="end"}`, nil)

	KeepaliveGoroutine     = NatSet.NewCounter(`natnet_keepalive_goroutine_total`)
	KeepaliveSwitch        = NatSet.NewCounter(`natnet_keepalive_switch_total`)
	KeepaliveConnErr       = NatSet.NewCounter(`natnet_keepalive_err_total{type="conn"}`)
	KeepaliveReadErr       = NatSet.NewCounter(`natnet_keepalive_err_total{type="read"}`)
	KeepaliveWriteErr      = NatSet.NewCounter(`natnet_keepalive_err_total{type="write"}`)
	KeepaliveReadBytes     = NatSet.NewCounter(`natnet_keepalive_bytes_total{type="read"}`)
	KeepaliveWriteBytes    = NatSet.NewCounter(`natnet_keepalive_bytes_total{type="write"}`)
	KeepaliveReadLastTime  = NatSet.NewGauge(`natnet_keepalive_last_time_unix{type="read",stage="end"}`, nil)
	KeepaliveWriteLastTime = NatSet.NewGauge(`natnet_keepalive_last_time_unix{type="write",stage="end"}`, nil)
)

func init() {
	metrics.RegisterSet(NatSet)
}
