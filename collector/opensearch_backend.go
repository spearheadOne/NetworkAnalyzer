package main

import (
	"bytes"
	"config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type OpenSearchBackend struct {
	client       *opensearch.Client
	flowIndex    string
	counterIndex string
}

func NewOpenSearchBackend(openSearchConfig config.OpenSearchConfig) (*OpenSearchBackend, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{openSearchConfig.Host},
	})
	if err != nil {
		return nil, err
	}

	return &OpenSearchBackend{client: client,
		flowIndex:    openSearchConfig.FlowIndex,
		counterIndex: openSearchConfig.CounterIndex}, nil
}

func (b *OpenSearchBackend) Index(events ParsedEvents) error {
	if len(events.Flows) == 0 && len(events.Counters) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var buf bytes.Buffer

	if err := writeBulkDocuments(&buf, b.flowIndex, events.Flows); err != nil {
		return fmt.Errorf("write counter bulk document: %w", err)
	}

	if err := writeBulkDocuments(&buf, b.counterIndex, events.Counters); err != nil {
		return fmt.Errorf("write counter bulk document: %w", err)
	}

	req := opensearchapi.BulkRequest{
		Body: bytes.NewReader(buf.Bytes()),
	}

	res, err := req.Do(ctx, b.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed index %q failed: %s", "index-name", res.Status())
	}

	return nil
}

func writeBulkDocuments[T any](buf *bytes.Buffer, indexName string, docs []T) error {
	for _, doc := range docs {
		if err := writeBulkDocument(buf, indexName, doc); err != nil {
			return err
		}
	}
	return nil
}

func writeBulkDocument(buf *bytes.Buffer, indexName string, doc any) error {
	if err := writeBulkLine(buf, map[string]interface{}{
		"index": map[string]interface{}{
			"_index": indexName,
		},
	}); err != nil {
		return err
	}

	if err := writeBulkLine(buf, doc); err != nil {
		return err
	}

	return nil
}

func writeBulkLine(buf *bytes.Buffer, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if _, err := buf.Write(raw); err != nil {
		return err
	}

	if err := buf.WriteByte('\n'); err != nil {
		return err
	}

	return nil
}
