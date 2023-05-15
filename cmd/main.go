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
	cset := NewClientSet(ctx, "default")

	timer := time.NewTicker(5 * time.Second)

	log.Printf("Application started successfully! Listening on port %s...", port)
	go srv.ListenAndServe()

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
