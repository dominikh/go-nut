package main

import (
	"net/http"

	"honnef.co/go/nut"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	c := nut.NewCollector([]string{"localhost"})
	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9999", nil)
}
