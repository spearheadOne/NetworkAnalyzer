SHELL := /bin/bash

COLLECTOR_DIR := collector

INDEXER_DIR := indexer

BUILD_DIR := bin

COMPOSE ?= docker compose

COMPOSE_FILE ?= docker-compose.yml


.PHONY: help
help:
	@echo "Targets:"
	@echo "  infra-up           Start OpenSearch infrastructure"
	@echo "  infra-down         Stop OpenSearch infrastructure"
	@echo "  infra-logs         Tail infrastructure logs"
	@echo ""
	@echo "  vm-up              Start Lima Mininet VM"
	@echo "  vm-down            Stop Lima Mininet VM"
	@echo "  vm-shell           Open shell in Lima Mininet VM"
	@echo "  topology-up        Run Mininet topology script"
	@echo "  traffic-help       Show iperf commands to run inside Mininet CLI"
	@echo ""
	@echo "  test               Run all tests"
	@echo "  test-collector     Run collector tests"
	@echo "  test-indexer       Run indexer tests"
	@echo ""
	@echo "  build              Build collector and indexer"
	@echo "  build-collector    Build collector binary"
	@echo "  build-indexer      Build indexer binary"
	@echo ""
	@echo "  run-collector      Run collector with ENV=local by default"
	@echo "  run-indexer        Run indexer; requires ACTION and INDEX"
	@echo ""
	@echo "  clean              Remove build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  ENV                Environment name (default: local)"
	@echo "  ACTION             Indexer action: create | delete | list"
	@echo "  INDEX              Indexer target: flow | counter | all"
	@echo "  COMPOSE            Docker Compose command (default: docker compose)"
	@echo "  COMPOSE_FILE       Compose file path (default: docker-compose.yml)"
	@echo ""
	@echo "Examples:"
	@echo "  make infra-up"
	@echo "  make test"
	@echo "  make build"
	@echo "  make run-collector ENV=local"
	@echo "  make run-indexer ENV=local ACTION=create INDEX=all"
	@echo "  make vm-up"
	@echo "  make topology-up"
	@echo "  make traffic-help"

.PHONY: infra-up

infra-up:

	$(COMPOSE) -f $(COMPOSE_FILE) up -d

.PHONY: infra-down

infra-down:

	$(COMPOSE) -f $(COMPOSE_FILE) down

.PHONY: infra-logs

infra-logs:

	$(COMPOSE) -f $(COMPOSE_FILE) logs -f

.PHONY: test

.PHONY: vm-up

vm-up:

	limactl start mininet-vm

.PHONY: vm-stop

vm-stop:

	limactl stop mininet-vm

.PHONY: topology-up

topology-up:

	cd mininet && sudo python3 topology.py

.PHONY: topology-up-exec

topology-up-exec:

	cd mininet && sudo ./topology.py

test: test-collector test-indexer

.PHONY: test-collector

test-collector:

	cd $(COLLECTOR_DIR) && go test -v ./...

.PHONY: test-indexer

test-indexer:

	cd $(INDEXER_DIR) && go test -v ./...

.PHONY: build

build: build-collector build-indexer

.PHONY: build-collector

build-collector:

	mkdir -p $(BUILD_DIR)
	cd $(COLLECTOR_DIR) && go build -o ../$(BUILD_DIR)/collector .

.PHONY: build-indexer

build-indexer:

	mkdir -p $(BUILD_DIR)
	cd $(INDEXER_DIR) && go build -o ../$(BUILD_DIR)/indexer .

ENV ?= local

.PHONY: run-collector

run-collector:

	cd $(COLLECTOR_DIR) && go run . -env=$(ENV)

ACTION ?=

INDEX ?=

.PHONY: run-indexer

run-indexer:

	@if [[ -z "$(ACTION)" || -z "$(INDEX)" ]]; then \
		echo "Usage: make run-indexer ENV=local ACTION=create INDEX=flow"; \
		exit 1; \
	fi
	cd $(INDEXER_DIR) && go run . -env=$(ENV) -action=$(ACTION) -index=$(INDEX)

.PHONY: clean

clean:

	rm -rf $(BUILD_DIR)

.PHONY: traffic-help

traffic-help:

	@echo "Run inside Mininet CLI:"
	@echo "  host2 iperf -s &"
	@echo "  host1 iperf -c 10.0.0.2 -t 20 -P 5"
	@echo "  host2 iperf -c 10.0.0.1 -t 20 -P 5"