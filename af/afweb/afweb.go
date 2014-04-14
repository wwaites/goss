package main

import (
	"flag"
	"fmt"
	"hubs.net.uk/oss/log"
	"hubs.net.uk/oss/rrd"
	"hubs.net.uk/oss/web"
	"net"
	"net/http/fcgi"
	"os"
)

var family string
var sock string
var prefix string
var data string
var debug bool
func init() {
        flag.Usage = func() {
                fmt.Fprintf(os.Stderr, `
Usage: %s [flags]

`, os.Args[0])
                flag.PrintDefaults()
        }
	flag.StringVar(&family, "family", "tcp", "FCGI socket address family")
        flag.StringVar(&sock, "sock", "127.0.0.1:6072", "FCGI socket address")
	flag.StringVar(&prefix, "prefix", "", "Request URL prefix")
	flag.StringVar(&data, "data", "/var/www/mrtg", "Data directori containing RRD files")
        flag.BoolVar(&debug, "d", false, "Debug")
}

func main() {
	flag.Parse()

	listener, err := net.Listen(family, sock)
	if err != nil {
		log.Err.Fatalf("net.Listen(\"%s\", \"%s\"): %s", family, sock, err)
	}

	mux := web.NewServeMux(prefix, "Foo Bar")

	grapher := rrd.NewGraphHandler(data)
	mux.Handle("/graph/", grapher)

	err = fcgi.Serve(listener, mux)
	if err != nil {
		log.Err.Fatalf("fcgi.Serve(): %s", err)
	}
}