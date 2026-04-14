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
		log.Fatalf("topology.host must not be empty in %q", cfg)
	}

	parser := &Parser{}
	collector := &Collector{cfg.Topology.Host, parser}
	collector.ListenUdp()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("./collector -env=local")
}
