// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

package whlog

import (
	"net"
	"net/http"
	"time"

	"github.com/jtolds/webhelp/whcompat"
)

const (
	keepAlivePeriod = 3 * time.Minute
)

// Serve takes a net.Listener, adds the TCPKeepAliveListener wrapper if
// possible, and serves incoming HTTP requests off of it.
func Serve(l net.Listener, handler http.Handler) error {
	if tcp_l, ok := l.(*net.TCPListener); ok {
		l = TCPKeepAliveListener(tcp_l)
	}
	return (&http.Server{Handler: handler}).Serve(l)
}

// TCPKeepAliveListener takes a *net.TCPListener and returns a net.Listener
// with TCP keep-alive semantics turned on.
func TCPKeepAliveListener(l *net.TCPListener) net.Listener {
	return tcpKeepAliveListener{TCPListener: l}
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(keepAlivePeriod)
	return tc, nil
}

// ListenAndServe creates a TCP listener prior to calling Serve. It also logs
// the address it listens on, and wraps given handlers in whcompat.DoneNotify.
func ListenAndServe(addr string, handler http.Handler) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	logger.Noticef("listening on %s", l.Addr())
	return Serve(l, whcompat.DoneNotify(handler))
}