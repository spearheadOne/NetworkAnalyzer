# Network Analyzer

A lightweight network monitoring system that simulates switch telemetry (sFlow) using Mininet and visualizes data via OpenSearch.

This project consists of:

* Mininet topology generating traffic and sFlow packets
* Collector (Go) receiving UDP sFlow data and indexing it
* Indexer (go) creating, deleting and listing indexes
* OpenSearch + Dashboards storing and visualizing metrics

## Architecture
Mininet (sFlow) → Collector (UDP) → OpenSearch → Dashboards

## Prerequisites

* Docker + Docker Compose
* Go (1.21+ recommended)
* Python 3 (for Mininet topology)
* Mininet (inside VM)
* Lima VM (for Mininet)

## Setup
1. Start infrastructure
```bash
make infra-up
```

2. Start Mininet VM
```bash
make vm-up
```

3. Run topology inside VM
```bash
sudo python3 topology.py
```

4. Create indexes
```bash
make run-indexer ACTION=create ENV=local
```
5. Start collector
```bash
make run-collector ENV=local
```

6. Generate traffic inside Mininet CLI inside Lima VM
```bash
host2 iperf -s &
host1 iperf -c 10.0.0.2 -t 20 -P 5
host2 iperf -c 10.0.0.1 -t 20 -P 5
```

### Verification
Check sFlow packets
```bash
sudo tcpdump -ni any udp port 6343
```

Check indexes
```bash
curl http://localhost:9200/_cat/indices?v
```

Check data
```bash
curl http://localhost:9200/sflow-flow-*/_search?pretty

curl http://localhost:9200/sflow-counter-*/_search?pretty
```
## Development
1. Run tests
```bash
make test
```
2. Run specfifc tests
```
make test-collector

make test-indexer
```
3. Build
```bash
make build
```

## Cleanup
1. Stop mininet inside VM
```bash
sudo mn-c
```
2. Stop vm
```aiignore
make vm-down
```

3. Stop infra
```aiignore
make infra-down
```

## Notes

* sFlow packets can contain multiple events → indexing load can be high
* Collector uses batching for efficiency
* Queue overflow will drop events (by design)


## Future Improvements

* Retry logic for failed indexing
* Backpressure handling instead of dropping events
* Metrics and monitoring for collector
* Predefined OpenSearch mappings


## Summary

This project demonstrates:

* network telemetry simulation (Mininet + sFlow)
* concurrent processing in Go
* real-time indexing into OpenSearch
* visualization via Dashboards