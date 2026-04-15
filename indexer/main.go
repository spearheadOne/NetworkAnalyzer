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
	indexFlag := flag.String("index", "", "Index: flow, counter, all")
	actionFlag := flag.String("action", "", "Action: create, list, delete")

	flag.Parse()

	if *envFlag == "" || *indexFlag == "" || *actionFlag == "" {
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

	backend, err := NewOpenSearchBackend(cfg.OpenSearch)
	if err != nil {
		log.Fatal(err)
	}

	indexer := NewIndexer(cfg.OpenSearch, backend)

	executor, err := NewExecutor(indexer, *actionFlag, *indexFlag)
	if err != nil {
		log.Fatal(err)
	}

	if err := executor.Execute(); err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  ./indexer -env=<env> -index=<index> -action=<action>")
	fmt.Println()

	fmt.Println("Available values:")
	fmt.Println("  env:    local | dev | uat | prod")
	fmt.Println("  index:  flow | counter | all")
	fmt.Println("  action: create | list | delete")
	fmt.Println()

	fmt.Println("Examples:")
	fmt.Println("  ./indexer -env=local -index=flow -action=create")
	fmt.Println("  ./indexer -env=local -index=all -action=delete")
}

func validateConfig(cfg *config.Config) error {
	if cfg.OpenSearch.Host == "" {
		return fmt.Errorf("opensearch.host must not be empty")
	}
	if cfg.OpenSearch.CounterIndex == "" {
		return fmt.Errorf("opensearch.counter-index must not be empty")
	}
	if cfg.OpenSearch.FlowIndex == "" {
		return fmt.Errorf("opensearch.flow-index must not be empty")
	}
	return nil
}
