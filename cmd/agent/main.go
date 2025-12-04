package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourname/metric-agent/internal/collector"
	"github.com/yourname/metric-agent/internal/server"
)

func main() {
	var (
		listenAddr = flag.String("listen", ":9100", "agent HTTP listen address")
		interval   = flag.Duration("interval", 5*time.Second, "collection interval")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// custom registry so we can mount to /metrics-prom
	reg := prometheus.NewRegistry()
	// create collector & sampler
	sampler := &collector.GopsSampler{}
	coll := collector.NewCollector(sampler, *interval, reg)

	// mount Prometheus handler on DefaultServeMux to expose at /metrics-prom
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.DefaultServeMux.Handle("/metrics-prom", promHandler)

	// start collector
	coll.Start(ctx)

	// start server
	srv := server.New(*listenAddr, coll)
	if err := srv.Start(); err != nil {
		fmt.Printf("server start error: %v\n", err)
		os.Exit(1)
	}

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
	coll.Stop()
}
