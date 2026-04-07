package store

import (
	"context"
	"fmt"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantStore manages connections to Qdrant and vector operations.
type QdrantStore struct {
	client *qdrant.Client
	config *StoreConfig
}

// StoreConfig contains Qdrant configuration.
type StoreConfig struct {
	Host       string        // Qdrant host (default: localhost)
	Port       uint32        // Qdrant port (default: 6333)
	APIKey     string        // API key (if using Qdrant cloud)
	Timeout    time.Duration // Request timeout
	MaxRetries int           // Max connection retries
}

// NewQdrantStore creates a new Qdrant store with retry logic.
func NewQdrantStore(ctx context.Context, config *StoreConfig) (*QdrantStore, error) {
	if config == nil {
		config = &StoreConfig{
			Host:       "localhost",
			Port:       6333,
			Timeout:    10 * time.Second,
			MaxRetries: 3,
		}
	}

	// Default values
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6333
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	// Create client with retries
	var client *qdrant.Client
	var lastErr error

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

		var err error
		client, err = qdrant.NewClient(&qdrant.Config{
			Addr:    addr,
			APIKey:  config.APIKey,
			Timeout: config.Timeout,
		})

		if err == nil {
			// Verify connection is working
			if err := client.HealthCheck(ctx); err == nil {
				return &QdrantStore{client: client, config: config}, nil
			}
			lastErr = err
		} else {
			lastErr = err
		}

		if attempt < config.MaxRetries-1 {
			time.Sleep(time.Second * time.Duration(attempt+1)) // Exponential backoff
		}
	}

	return nil, fmt.Errorf("failed to connect to Qdrant after %d retries: %w", config.MaxRetries, lastErr)
}

// Close closes the Qdrant client connection.
func (s *QdrantStore) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// CreateCollection creates a vector collection in Qdrant.
func (s *QdrantStore) CreateCollection(ctx context.Context, collectionName string, vectorSize uint64) error {
	if collectionName == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	if vectorSize == 0 {
		vectorSize = 384 // Default FastEmbed model vector size
	}

	// Check if collection already exists
	exists, err := s.collectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if exists {
		return nil // Collection already exists
	}

	// Create collection
	err = s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})

	if err != nil {
		return fmt.Errorf("failed to create collection %s: %w", collectionName, err)
	}

	return nil
}

// UpsertVector adds or updates a vector in the collection.
func (s *QdrantStore) UpsertVector(ctx context.Context, collectionName string, pointID uint64, vector []float32, payload map[string]interface{}) error {
	if collectionName == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if len(vector) == 0 {
		return fmt.Errorf("vector cannot be empty")
	}

	// Convert payload to qdrant StructValue
	payloadData := make(map[string]*qdrant.Value)
	for key, val := range payload {
		payloadData[key] = &qdrant.Value{
			Kind: &qdrant.Value_StringValue{
				StringValue: fmt.Sprintf("%v", val),
			},
		}
	}

	points := []*qdrant.PointStruct{
		{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Num{
					Num: pointID,
				},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: vector,
					},
				},
			},
			Payload: payloadData,
		},
	}

	err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	if err != nil {
		return fmt.Errorf("failed to upsert vector: %w", err)
	}

	return nil
}

// QueryVectors searches for similar vectors in a collection.
func (s *QdrantStore) QueryVectors(ctx context.Context, collectionName string, vector []float32, limit uint64, scoreThreshold float32) ([]*qdrant.ScoredPoint, error) {
	if collectionName == "" {
		return nil, fmt.Errorf("collection name cannot be empty")
	}
	if len(vector) == 0 {
		return nil, fmt.Errorf("query vector cannot be empty")
	}
	if limit == 0 {
		limit = 10 // Default limit
	}

	results, err := s.client.Search(ctx, &qdrant.SearchPoints{
		CollectionName: collectionName,
		Vector:         vector,
		Limit:          limit,
		ScoreThreshold: &scoreThreshold,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{
				Enable: &qdrant.Boolean{Bool: true},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query vectors: %w", err)
	}

	return results, nil
}

// DeleteCollection deletes a collection from Qdrant.
func (s *QdrantStore) DeleteCollection(ctx context.Context, collectionName string) error {
	err := s.client.DeleteCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
	}
	return nil
}

// GetCollectionInfo retrieves metadata about a collection.
func (s *QdrantStore) GetCollectionInfo(ctx context.Context, collectionName string) (*qdrant.CollectionInfo, error) {
	info, err := s.client.GetCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info for %s: %w", collectionName, err)
	}
	return info, nil
}

// HealthCheck verifies the Qdrant service is available.
func (s *QdrantStore) HealthCheck(ctx context.Context) error {
	return s.client.HealthCheck(ctx)
}

// collectionExists checks if a collection exists.
func (s *QdrantStore) collectionExists(ctx context.Context, collectionName string) (bool, error) {
	_, err := s.client.GetCollection(ctx, collectionName)
	if err != nil {
		// Check for "not found" style errors
		if fmt.Sprintf("%v", err) == "Not Found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
