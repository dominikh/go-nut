package main

import (
	"net/http"

	"honnef.co/go/nut/nutcollector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	c := nutcollector.New([]string{"localhost"})
	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9999", nil)
}
