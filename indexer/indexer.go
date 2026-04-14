package main

import "github.com/opensearch-project/opensearch-go"

type Indexer struct {
	client *opensearch.Client
}

func (i *Indexer) CreateIndex() {

}

func (i *Indexer) DeleteIndex() {

}
