package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const (
	port string = "9090"
)

var (
	// Initialise the prometheus gauge metric to monitor health
	healthcheck = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "up",
			Help: "Service healthcheck",
		},
		[]string{"valid"},
	)
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	srv := NewMetricServer(ctx, port)
	cset := NewClientSet(ctx, "default")

	timer := time.NewTicker(5 * time.Second)

	log.Printf("Application started successfully! Listening on port %s...", port)
	go srv.ListenAndServe()

	go func() {
		for range timer.C {
			fmt.Println("Timer ticked!")
			cset.countPods()
		}
	}()

	// Use interrupt signal channel to block program from exiting immediately
	// Only proceeds when an interrupt signal (Ctrl+C) writes a msg to the channel
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("Application closing...")
	srv.Close()
	timer.Stop()
	defer cancel()
}
