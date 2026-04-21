package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOpenSearchBackend_Index(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "opensearchproject/opensearch:latest",
		ExposedPorts: []string{"9200/tcp"},
		Env: map[string]string{
			"discovery.type":          "single-node",
			"DISABLE_SECURITY_PLUGIN": "true",
		},
		WaitingFor: wait.ForHTTP("/").
			WithPort("9200/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "9200")
	require.NoError(t, err)

	addr := fmt.Sprintf("http://%s:%s", host, port.Port())

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{addr},
		Transport: &http.Transport{},
	})
	require.NoError(t, err)

	backend := &OpenSearchBackend{client: client, flowIndex: "test-flow", counterIndex: "test-counter"}
	err = createIndexes(backend, ctx)
	require.NoError(t, err)

	events := ParsedEvents{

		Flows: []FlowEvent{
			{
				Event: Event{
					Timestamp: time.Now().UTC(),
					Kind:      "flow",
					AgentIP:   "192.168.5.15",
					Collector: "collector-mac",
				},
				FrameLength: 128,
				SampleRate:  1,
			},
		},
		Counters: []CounterEvent{
			{
				Event: Event{
					Timestamp: time.Now().UTC(),
					Kind:      "counter",
					AgentIP:   "192.168.5.15",
					Collector: "collector-mac",
				},
				IfIndex:      2,
				InOctets:     1000,
				OutOctets:    2000,
				InUcastPkts:  10,
				OutUcastPkts: 20,
			},
		},
	}

	err = backend.Index(events)
	require.NoError(t, err)

	err = refreshIndex(client, []string{"test-flow", "test-counter"}, ctx)
	require.NoError(t, err)

	flowCount, err := countDocuments(client, "test-flow", ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), flowCount)
	counterCount, err := countDocuments(client, "test-counter", ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), counterCount)

}

func TestOpenSearchBackend_Index_EmptyEvents(t *testing.T) {
	backend := &OpenSearchBackend{
		client:       nil,
		flowIndex:    "test-flow",
		counterIndex: "test-counter",
	}
	err := backend.Index(ParsedEvents{})
	require.NoError(t, err)

}

func createIndexes(backend *OpenSearchBackend, ctx context.Context) error {
	if err := createIndex(backend.client, backend.flowIndex, ctx); err != nil {
		return err
	}

	if err := createIndex(backend.client, backend.counterIndex, ctx); err != nil {
		return err
	}

	return nil
}

func createIndex(client *opensearch.Client, indexName string, ctx context.Context) error {
	req := opensearchapi.IndicesCreateRequest{
		Index: indexName,
	}

	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index %q failed: %s", indexName, res.Status())
	}

	return nil
}

func countDocuments(client *opensearch.Client, indexName string, ctx context.Context) (int64, error) {
	req := opensearchapi.CountRequest{
		Index: []string{indexName},
	}
	res, err := req.Do(ctx, client)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.IsError() {
		return 0, fmt.Errorf("count request for index %q failed: %s", indexName, res.Status())
	}
	var response struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return 0, err
	}
	return response.Count, nil

}

func refreshIndex(client *opensearch.Client, indexes []string, ctx context.Context) error {

	req := opensearchapi.IndicesRefreshRequest{
		Index: indexes,
	}
	res, err := req.Do(ctx, client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("refresh failed: %s", res.Status())
	}
	return nil

}
