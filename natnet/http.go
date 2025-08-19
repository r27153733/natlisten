package natnet

import (
	"errors"
	"net"
	"net/http"
	"syscall"
)

func HttpReuseListenAndServe(s *http.Server) error {
	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	lnConn, ok := ln.(syscall.Conn)
	if !ok {
		return errors.New("could not convert to syscall.Conn")
	}
	err = SetListenerReuse(lnConn)
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

func ListenReuseAndServeTLS(s *http.Server, certFile, keyFile string) error {
	addr := s.Addr
	if addr == "" {
		addr = ":https"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer ln.Close()

	lnConn, ok := ln.(syscall.Conn)
	if !ok {
		return errors.New("could not convert to syscall.Conn")
	}
	err = SetListenerReuse(lnConn)
	if err != nil {
		return err
	}

	return s.ServeTLS(ln, certFile, keyFile)
}
