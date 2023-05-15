package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	port string = "9090"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	srv := NewMetricServer(ctx, port)
	// create client set to scrape from all namespaces, leave namespace ""
	// improve by allowing this to be externally configurable
	cset := NewClientSet(ctx, "")

	// scrape every pod info every 5 seconds.
	// improve by allowing this to be externally configurable
	timer := time.NewTicker(5 * time.Second)

	// Launch metrics server
	go srv.ListenAndServe()
	log.Printf("Application started successfully! Listening on port %s...", port)

	// Launch pod scraper
	go func() {
		for range timer.C {
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
