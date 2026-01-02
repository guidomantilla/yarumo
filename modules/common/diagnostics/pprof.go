package diagnostics

import (
	"net/http"
	"net/http/pprof"
)

func NewPprofHandler() http.Handler {
	debugHandler := http.NewServeMux()
	debugHandler.HandleFunc("/debug/pprof/", pprof.Index)
	debugHandler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugHandler.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugHandler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugHandler.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return debugHandler
}
