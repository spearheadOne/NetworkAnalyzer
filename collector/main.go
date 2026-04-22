package main

import (
	"collector/opensearch"
	"config"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	envFlag := flag.String("env", "", "Environment: local,dev,uat,prod")
	flag.Parse()

	if *envFlag == "" {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load(*envFlag)
	if err != nil {
		log.Fatal(err)
	}

	if err = validateConfig(cfg); err != nil {
		log.Fatal(err)
	}

	parser := &Parser{}
	indexBackend, err := opensearch.NewOpenSearchBackend(cfg.OpenSearch)
	if err != nil {
		log.Fatal(err)
	}

	collector := NewCollector(cfg.Topology.Addr, parser, indexBackend, cfg.Collector.QueueSize, cfg.Collector.WorkersNum)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err = collector.Start(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("collector started")

	collector.Wait()
	log.Println("collector stopped")
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("./collector -env=local")
}

func validateConfig(cfg *config.Config) error {

	if cfg.Topology.Addr == "" {
		return fmt.Errorf("topology.host must not be empty")
	}
	if cfg.OpenSearch.Host == "" {
		return fmt.Errorf("opensearch.host must not be empty")
	}
	if cfg.OpenSearch.FlowIndex == "" {
		return fmt.Errorf("opensearch.flow-index must not be empty")
	}
	if cfg.OpenSearch.CounterIndex == "" {
		return fmt.Errorf("opensearch.counter-index must not be empty")
	}
	if cfg.Collector.WorkersNum <= 0 {
		return fmt.Errorf("collector.workers-num must be > 0")
	}
	if cfg.Collector.QueueSize <= 0 {
		return fmt.Errorf("collector.queue-size must be > 0")
	}
	return nil

}
