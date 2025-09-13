package nathttp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"syscall"

	"github.com/r27153733/natlisten/natnet"
)

func ReuseListenAndServeIPV4PubNat(ctx context.Context, s *http.Server, fn func(ip net.IP, port int) error) error {
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
	err = natnet.SetListenerReuse(lnConn)
	if err != nil {
		return err
	}
	localAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return errors.New("could not convert to net.TCPAddr")
	}
	err = natnet.IPV4PubNat(ctx, localAddr, fn)
	if err != nil {
		return err
	}

	return s.Serve(ln)
}

func ReuseListenAndServeTLSIPV4PubNat(ctx context.Context, s *http.Server, certFile, keyFile string, fn func(ip net.IP, port int) error) error {
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
	err = natnet.SetListenerReuse(lnConn)
	if err != nil {
		return err
	}

	localAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return errors.New("could not convert to net.TCPAddr")
	}
	err = natnet.IPV4PubNat(ctx, localAddr, fn)
	if err != nil {
		return err
	}

	return s.ServeTLS(ln, certFile, keyFile)
}

func ReuseListenAndServeTLSConfigIPV4PubNat(ctx context.Context, s *http.Server, config *tls.Config, fn func(ip net.IP, port int) error) error {
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
	err = natnet.SetListenerReuse(lnConn)
	if err != nil {
		return err
	}

	localAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return errors.New("could not convert to net.TCPAddr")
	}
	err = natnet.IPV4PubNat(ctx, localAddr, fn)
	if err != nil {
		return err
	}

	httpsLn := tls.NewListener(ln, config)

	return s.Serve(httpsLn)
}

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
	err = natnet.SetListenerReuse(lnConn)
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
	err = natnet.SetListenerReuse(lnConn)
	if err != nil {
		return err
	}

	return s.ServeTLS(ln, certFile, keyFile)
}
