package main

import (
	"config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type OpenSearchBackend struct {
	client *opensearch.Client
}

func NewOpenSearchBackend(openSearchConfig config.OpenSearchConfig) (*OpenSearchBackend, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{openSearchConfig.Host},
	})
	if err != nil {
		return nil, err
	}

	return &OpenSearchBackend{client: client}, nil
}

func (b *OpenSearchBackend) CreateIndex(indexName string) error {
	ctx, cancel := createContext()
	defer cancel()

	req := opensearchapi.IndicesCreateRequest{
		Index: indexName,
	}

	res, err := req.Do(ctx, b.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index %q failed: %s", indexName, res.Status())
	}

	return nil
}

func (b *OpenSearchBackend) DeleteIndexes(indexes []string) error {
	ctx, cancel := createContext()
	defer cancel()

	req := opensearchapi.IndicesDeleteRequest{
		Index: indexes,
	}

	res, err := req.Do(ctx, b.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete index %q failed: %s", indexes, res.Status())
	}

	return nil
}

func (b *OpenSearchBackend) ListIndexes() ([]string, error) {
	ctx, cancel := createContext()
	defer cancel()

	req := opensearchapi.CatIndicesRequest{
		Format: "json",
	}

	res, err := req.Do(ctx, b.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("list indices failed: %s", res.Status())
	}

	var data []struct {
		Index string `json:"index"`
	}

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	indexes := make([]string, 0, len(data))

	for _, item := range data {
		indexes = append(indexes, item.Index)
	}

	return indexes, nil
}

func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 20*time.Second)
}
