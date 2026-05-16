package diagnostics

import (
	"net/http"
	nhpprof "net/http/pprof"
)

// NewPprofHandler returns a handler that exposes pprof endpoints.
func NewPprofHandler() http.Handler {
	debugHandler := http.NewServeMux()
	debugHandler.HandleFunc("/debug/pprof/", nhpprof.Index)
	debugHandler.HandleFunc("/debug/pprof/cmdline", nhpprof.Cmdline)
	debugHandler.HandleFunc("/debug/pprof/profile", nhpprof.Profile)
	debugHandler.HandleFunc("/debug/pprof/symbol", nhpprof.Symbol)
	debugHandler.HandleFunc("/debug/pprof/trace", nhpprof.Trace)

	return debugHandler
}
