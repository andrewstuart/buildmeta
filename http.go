package buildmeta

import (
	"net/http"
	"path"

	health "astuart.co/go-healthcheck"
)

// Well-known paths used by buildmeta
const (
	BasePath  = "/-/"
	InfoPath  = "info"
	AlivePath = "alive"
	PingPath  = "ping"
	ReadyPath = "ready"
)

// Handler returns a MetaHandler, which serves up version information and
// any registered health checks at given endpoints. Liveness and Readiness
// probes should be registered in Alive and Ready.
func Handler() *MetaHandler {
	return &MetaHandler{
		Info:  GetInfo(),
		Alive: health.NewRegistry(),
		Ready: health.NewRegistry(),
	}
}

// MetaHandler serves structured metadata for version control metadata,
// liveness, and readiness.
type MetaHandler struct {
	Info         Info
	Alive, Ready *health.Registry
}

// The MetaHandler implements http.ServeMux by serving both the info struct and
// health check information.
func (rh *MetaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch path.Base(r.URL.Path) {
	case InfoPath:
		rh.Info.ServeHTTP(w, r)
	case AlivePath:
		rh.Alive.ServeHTTP(w, r)
	case PingPath:
		w.Write([]byte("pong"))
	case ReadyPath:
		rh.Ready.ServeHTTP(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
