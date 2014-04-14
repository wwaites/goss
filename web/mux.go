package web

import (
	"hubs.net.uk/oss/log"
	"net/http"
	"strings"
)

type ServeMux struct {
	*http.ServeMux
	prefix string
	name string
}

/*
 * prefix is the on which this mux is mounted.
 * name is an informative name for this server software.
 */
func NewServeMux(prefix, name string) (*ServeMux) {
	return &ServeMux{http.NewServeMux(), prefix, name}
}

/*
 * A wrapper around http.ServMux that takes care of logging and setting
 * the Server header, etc. It also strips the configured prefix to
 * facilitate mounting as a FCGI server.
 */
func (s *ServeMux) ServeHTTP (w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", s.name)
	log.Info.Printf("%s %s", r.RemoteAddr, r.URL)
	r.URL.Path = strings.TrimPrefix(r.URL.Path, s.prefix)
	s.ServeMux.ServeHTTP(w, r)
}
