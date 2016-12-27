// Copyright (C) 2016 JT Olds
// See LICENSE for copying information

// Package whgls provides webhelp tools that use grossness enabled by
// the github.com/jtolds/gls package. No other webhelp packages use
// github.com/jtolds/gls.
//
// The predominant use case for github.com/jtolds/gls is to attach a current
// request's contextual information to all log lines kicked off by the request.
package whgls

import (
	"log"
	"net/http"

	"github.com/jtolds/gls"
	"github.com/jtolds/webhelp/whcompat"
	"github.com/jtolds/webhelp/whroute"
	"golang.org/x/net/context"
)

var (
	ctxMgr = gls.NewContextManager()
	reqSym = gls.GenSym()
)

// Bind will make sure that Load works from any (reasonably shallow)
// callstacks kicked off by this handler, via the magic of
// github.com/jtolds/gls.
func Bind(h http.Handler) http.Handler {
	return whroute.HandlerFunc(h, func(w http.ResponseWriter, r *http.Request) {
		ctxMgr.SetValues(gls.Values{reqSym: r}, func() {
			h.ServeHTTP(w, r)
		})
	})
}

// Load will return the *http.Request bound to the current call stack by a
// Bind handler further up the stack.
func Load() *http.Request {
	if val, ok := ctxMgr.GetValue(reqSym); ok {
		if r, ok := val.(*http.Request); ok {
			return r
		}
	}
	return nil
}

// CtxLogger is a logger that requires a request context to work.
type CtxLogger func(ctx context.Context, format string, args ...interface{})

// SetLogOutput will configure the standard library's logger to use the
// provided logger that requires a context, such as AppEngine's loggers.
// This requires that the handler was wrapped with Bind. Note that this will
// cause all log messages without a context (including ones from deep
// callstacks due to github.com/jtolds/gls limitations) to be silently
// swallowed!
//
// The benefit of this is that the standard library's logger (or some other
// logger that doesn't use contexts) can now be used naturally on a platform
// that requires contexts (like App Engine).
//
// App Engine Example:
//
//  import (
//    "net/http"
//
//    "github.com/jtolds/webhelp/whgls"
//    "google.golang.org/appengine/log"
//  )
//
//  var (
//    handler = ...
//  )
//
//  func init() {
//    whgls.SetLogOutput(log.Infof)
//    http.Handle("/", whgls.Bind(handler))
//  }
//
func SetLogOutput(logger CtxLogger) {
	log.SetOutput(writerFunc(func(p []byte) (n int, err error) {
		r := Load()
		if r == nil {
			return len(p), nil
		}
		logger(whcompat.Context(r), "%s", string(p))
		return len(p), nil
	}))
}

type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(p []byte) (int, error) { return w(p) }