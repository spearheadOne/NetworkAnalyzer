package main

import (
	"config"
	"log"
)

type IndexBackend interface {
	CreateIndex(indexName string) error
	DeleteIndexes(indexes []string) error
	ListIndexes() ([]string, error)
}

type Indexer struct {
	backend      IndexBackend
	flowIndex    string
	counterIndex string
}

func NewIndexer(config config.OpenSearchConfig, backend IndexBackend) *Indexer {
	return &Indexer{
		backend:      backend,
		flowIndex:    config.FlowIndex,
		counterIndex: config.CounterIndex,
	}
}

func (i *Indexer) CreateFlowIndex() error {
	return i.backend.CreateIndex(i.flowIndex)
}

func (i *Indexer) CreateCounterIndex() error {
	return i.backend.CreateIndex(i.counterIndex)
}

func (i *Indexer) CreateIndexes() error {
	if err := i.CreateFlowIndex(); err != nil {
		return err
	}

	if err := i.CreateCounterIndex(); err != nil {
		return err
	}

	return nil
}

func (i *Indexer) DeleteFlowIndex() error {
	return i.backend.DeleteIndexes([]string{i.flowIndex})
}

func (i *Indexer) DeleteCounterIndex() error {
	return i.backend.DeleteIndexes([]string{i.counterIndex})
}

func (i *Indexer) DeleteAllIndexes() error {
	return i.backend.DeleteIndexes([]string{i.flowIndex, i.counterIndex})
}

func (i *Indexer) ListIndexes() error {
	indexes, err := i.backend.ListIndexes()
	if err != nil {
		return err
	}

	log.Println("Indexes:")
	for _, index := range indexes {
		log.Println(index)
	}

	return nil
}
