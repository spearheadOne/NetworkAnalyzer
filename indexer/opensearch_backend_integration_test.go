package main

import (
	"context"
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

func TestOpenSearchBackend_CreateAndDeleteIndex(t *testing.T) {
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

	backend := &OpenSearchBackend{client: client}
	indexName := "test-index"

	err = backend.CreateIndex(indexName)
	require.NoError(t, err)

	exists, err := indexExists(client, indexName)
	require.NoError(t, err)
	require.True(t, exists)

	err = backend.DeleteIndexes([]string{indexName})
	require.NoError(t, err)

	exists, err = indexExists(client, indexName)
	require.NoError(t, err)
	require.False(t, exists)
}

func indexExists(client *opensearch.Client, index string) (bool, error) {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	}
	if res.StatusCode == 404 {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status: %s", res.Status())
}
