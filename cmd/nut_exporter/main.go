// Command nut_exporter exports UPSs on one or more NUT servers to
// Prometheus.
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"honnef.co/go/nut/nutcollector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	fHosts := flag.String("h", "localhost", "Space-separated list of hosts to collect")
	fListen := flag.String("l", ":9100", "Address and port to listen on")
	flag.Parse()

	c := nutcollector.New(strings.Fields(*fHosts))
	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*fListen, nil))
}
