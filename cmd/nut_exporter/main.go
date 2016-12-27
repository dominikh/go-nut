package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"honnef.co/go/nut_exporter"
)

func main() {
	c := nut_exporter.NewNUTCollector([]string{"main"}, nil)
	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9999", nil)
}
