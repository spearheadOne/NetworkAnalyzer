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

	if cfg.OpenSearch.Host == "" {
		log.Fatalf("opensearch.host must not be empty in %q", cfg.OpenSearch)
	}
	if cfg.OpenSearch.CounterIndex == "" {
		log.Fatalf("opensearch.counter-index must not be empty in %q", cfg.OpenSearch)
	}

	if cfg.OpenSearch.FlowIndex == "" {
		log.Fatalf("opensearch.flow-index must not be empty in %q", cfg.OpenSearch)
	}

	backend, err := NewOpenSearchBackend(cfg.OpenSearch)
	if err != nil {
		log.Fatal(err)
	}

	indexer := NewIndexer(cfg.OpenSearch, backend)
	indexer.CreateFlowIndex()

}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("./indexer -env=local")
}
