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

	env, err := config.ParseEnvironment(*envFlag)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Load(env)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.OpenSearch.Host)

}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("./indexer -env=local")
}
