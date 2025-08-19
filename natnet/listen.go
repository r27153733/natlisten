package natnet

import (
	"syscall"
)

func SetListenerReuse(listener syscall.Conn) error {
	sysListener, err := listener.SyscallConn()
	if err != nil {
		return err
	}
	return sysListener.Control(setReuse)
}
