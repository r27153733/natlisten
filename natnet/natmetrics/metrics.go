package natmetrics

import "github.com/r27153733/natlisten/natnet/internal/metricset"

func GetStunProbeGoroutineCount() uint64 {
	return metricset.StunProbeGoroutine.Get()
}

func GetStunProbeSwitchCount() uint64 {
	return metricset.StunProbeSwitch.Get()
}

func GetStunProbeConnErrCount() uint64 {
	return metricset.StunProbeConnErr.Get()
}

func GetStunProbeSTUNErrCount() uint64 {
	return metricset.StunProbeSTUNErr.Get()
}

func GetStunProbeCallbackErrCount() uint64 {
	return metricset.StunProbeCallbackErr.Get()
}

func GetKeepaliveGoroutineCount() uint64 {
	return metricset.KeepaliveGoroutine.Get()
}

func GetKeepaliveSwitchCount() uint64 {
	return metricset.KeepaliveSwitch.Get()
}

func GetKeepaliveConnErrCount() uint64 {
	return metricset.KeepaliveConnErr.Get()
}

func GetKeepaliveReadErrCount() uint64 {
	return metricset.KeepaliveReadErr.Get()
}

func GetKeepaliveWriteErrCount() uint64 {
	return metricset.KeepaliveWriteErr.Get()
}

func GetKeepaliveReadBytesCount() uint64 {
	return metricset.KeepaliveReadBytes.Get()
}

func GetKeepaliveWriteBytesCount() uint64 {
	return metricset.KeepaliveWriteBytes.Get()
}

func GetStunProbeCallbackLastTime() float64 {
	return metricset.StunProbeCallbackLastTime.Get()
}

func GetKeepaliveReadLastTime() float64 {
	return metricset.KeepaliveReadLastTime.Get()
}

func GetKeepaliveWriteLastTime() float64 {
	return metricset.KeepaliveWriteLastTime.Get()
}
