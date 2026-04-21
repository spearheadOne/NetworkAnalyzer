package main

import (
	"config"
	"flag"
	"fmt"
	"log"
	"os"
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

	if cfg.Topology.Host == "" {
		log.Fatalf("topology.host must not be empty")
	}

	parser := &Parser{}
	indexBackend, err := NewOpenSearchBackend(cfg.OpenSearch)
	if err != nil {
		log.Fatal(err)
	}

	indexer := &Writer{indexBackend}
	collector := &Collector{cfg.Topology.Host, parser, indexer}
	collector.ListenUdp()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("./collector -env=local")
}
