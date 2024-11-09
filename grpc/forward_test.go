package grpc

import (
	"context"
	"github.com/qdrant/go-client/qdrant"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestForwardServerStart(t *testing.T) {
	ForwardServerStart()
}

func TestQDRantClient(t *testing.T) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 5000,
	})
	require.NoError(t, err)
	healthCheck, err := client.HealthCheck(context.Background())
	require.NoError(t, err)
	t.Log(healthCheck)
}

func TestGinClient(t *testing.T) {
	response, err := http.Get("http://127.0.0.1:5000/")
	require.NoError(t, err)
	t.Log(response.StatusCode)
}
