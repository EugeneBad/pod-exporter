package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type MetricServer struct {
	server  *http.Server
	context context.Context
}

// Return server object that listens on confgured port

func NewMetricServer(ctx context.Context, port string) *MetricServer {
	return &MetricServer{
		server:  &http.Server{Addr: fmt.Sprintf(":%s", port)},
		context: ctx,
	}
}

func (srv *MetricServer) ListenAndServe() {
	http.Handle("/metrics", promhttp.Handler())
	srv.server.ListenAndServe()
}

// Gracefully shutsdown server by closing the main context.
func (srv *MetricServer) Close() {
	log.Info("Graceful server shutdown!")
	if err := srv.server.Shutdown(srv.context); err != nil {
		log.Fatalf("Graceful server shutdown failed:%+s", err)
	}
}
