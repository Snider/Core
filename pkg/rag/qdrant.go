// Package rag provides RAG (Retrieval Augmented Generation) functionality
// for storing and querying documentation in Qdrant vector database.
package rag

import (
	"context"
	"fmt"

	"github.com/host-uk/core/pkg/log"
	"github.com/qdrant/go-client/qdrant"
)

// QdrantConfig holds Qdrant connection configuration.
type QdrantConfig struct {
	Host   string
	Port   int
	APIKey string
	UseTLS bool
}

// DefaultQdrantConfig returns default Qdrant configuration.
// Host defaults to localhost for local development.
func DefaultQdrantConfig() QdrantConfig {
	return QdrantConfig{
		Host:   "localhost",
		Port:   6334, // gRPC port
		UseTLS: false,
	}
}

// QdrantClient wraps the Qdrant Go client with convenience methods.
type QdrantClient struct {
	client *qdrant.Client
	config QdrantConfig
}

// NewQdrantClient creates a new Qdrant client.
func NewQdrantClient(cfg QdrantConfig) (*QdrantClient, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Host,
		Port:   cfg.Port,
		APIKey: cfg.APIKey,
		UseTLS: cfg.UseTLS,
	})
	if err != nil {
		return nil, log.E("rag.Qdrant", fmt.Sprintf("failed to connect to Qdrant at %s", addr), err)
	}

	return &QdrantClient{
		client: client,
		config: cfg,
	}, nil
}

// Close closes the Qdrant client connection.
func (q *QdrantClient) Close() error {
	return q.client.Close()
}

// HealthCheck verifies the connection to Qdrant.
func (q *QdrantClient) HealthCheck(ctx context.Context) error {
	_, err := q.client.HealthCheck(ctx)
	return err
}

// ListCollections returns all collection names.
func (q *QdrantClient) ListCollections(ctx context.Context) ([]string, error) {
	resp, err := q.client.ListCollections(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(resp))
	copy(names, resp)
	return names, nil
}

// CollectionExists checks if a collection exists.
func (q *QdrantClient) CollectionExists(ctx context.Context, name string) (bool, error) {
	return q.client.CollectionExists(ctx, name)
}

// CreateCollection creates a new collection with cosine distance.
func (q *QdrantClient) CreateCollection(ctx context.Context, name string, vectorSize uint64) error {
	return q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
}

// DeleteCollection deletes a collection.
func (q *QdrantClient) DeleteCollection(ctx context.Context, name string) error {
	return q.client.DeleteCollection(ctx, name)
}

// CollectionInfo returns information about a collection.
func (q *QdrantClient) CollectionInfo(ctx context.Context, name string) (*qdrant.CollectionInfo, error) {
	return q.client.GetCollectionInfo(ctx, name)
}

// Point represents a vector point with payload.
type Point struct {
	ID      string
	Vector  []float32
	Payload map[string]any
}

// UpsertPoints inserts or updates points in a collection.
func (q *QdrantClient) UpsertPoints(ctx context.Context, collection string, points []Point) error {
	if len(points) == 0 {
		return nil
	}

	qdrantPoints := make([]*qdrant.PointStruct, len(points))
	for i, p := range points {
		qdrantPoints[i] = &qdrant.PointStruct{
			Id:      qdrant.NewID(p.ID),
			Vectors: qdrant.NewVectors(p.Vector...),
			Payload: qdrant.NewValueMap(p.Payload),
		}
	}

	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collection,
		Points:         qdrantPoints,
	})
	return err
}

// SearchResult represents a search result with score.
type SearchResult struct {
	ID      string
	Score   float32
	Payload map[string]any
}

// Search performs a vector similarity search.
func (q *QdrantClient) Search(ctx context.Context, collection string, vector []float32, limit uint64, filter map[string]string) ([]SearchResult, error) {
	query := &qdrant.QueryPoints{
		CollectionName: collection,
		Query:          qdrant.NewQuery(vector...),
		Limit:          qdrant.PtrOf(limit),
		WithPayload:    qdrant.NewWithPayload(true),
	}

	// Add filter if provided
	if len(filter) > 0 {
		conditions := make([]*qdrant.Condition, 0, len(filter))
		for k, v := range filter {
			conditions = append(conditions, qdrant.NewMatch(k, v))
		}
		query.Filter = &qdrant.Filter{
			Must: conditions,
		}
	}

	resp, err := q.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(resp))
	for i, p := range resp {
		payload := make(map[string]any)
		for k, v := range p.Payload {
			payload[k] = valueToGo(v)
		}
		results[i] = SearchResult{
			ID:      pointIDToString(p.Id),
			Score:   p.Score,
			Payload: payload,
		}
	}
	return results, nil
}

// pointIDToString converts a Qdrant point ID to string.
func pointIDToString(id *qdrant.PointId) string {
	if id == nil {
		return ""
	}
	switch v := id.PointIdOptions.(type) {
	case *qdrant.PointId_Num:
		return fmt.Sprintf("%d", v.Num)
	case *qdrant.PointId_Uuid:
		return v.Uuid
	default:
		return ""
	}
}

// valueToGo converts a Qdrant value to a Go value.
func valueToGo(v *qdrant.Value) any {
	if v == nil {
		return nil
	}
	switch val := v.Kind.(type) {
	case *qdrant.Value_StringValue:
		return val.StringValue
	case *qdrant.Value_IntegerValue:
		return val.IntegerValue
	case *qdrant.Value_DoubleValue:
		return val.DoubleValue
	case *qdrant.Value_BoolValue:
		return val.BoolValue
	case *qdrant.Value_ListValue:
		list := make([]any, len(val.ListValue.Values))
		for i, item := range val.ListValue.Values {
			list[i] = valueToGo(item)
		}
		return list
	case *qdrant.Value_StructValue:
		m := make(map[string]any)
		for k, item := range val.StructValue.Fields {
			m[k] = valueToGo(item)
		}
		return m
	default:
		return nil
	}
}
