package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// QdrantManager manages Qdrant vector database
type QdrantManager struct {
	endpoint   string
	logChannel chan string
}

// QdrantStatus represents Qdrant health status
type QdrantStatus struct {
	IsHealthy   bool     `json:"is_healthy"`
	Collections []string `json:"collections"`
}

// NewQdrantManager creates a new qdrant manager
func NewQdrantManager(endpoint string, logCh chan string) *QdrantManager {
	return &QdrantManager{
		endpoint:   endpoint,
		logChannel: logCh,
	}
}

// HealthCheck verifies Qdrant is running and healthy
func (m *QdrantManager) HealthCheck(ctx context.Context) (bool, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(m.endpoint + "/health")
	if err != nil {
		m.log(fmt.Sprintf("Health check failed: %v", err))
		return false, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		m.log(fmt.Sprintf("Health check returned status %d", resp.StatusCode))
		return false, fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	
	m.log("Health check passed")
	return true, nil
}

// EnsureCollections creates required collections if they don't exist
func (m *QdrantManager) EnsureCollections(ctx context.Context) error {
	requiredCollections := []string{"memory", "skills", "rules"}
	
	for _, collName := range requiredCollections {
		exists, err := m.collectionExists(ctx, collName)
		if err != nil {
			m.log(fmt.Sprintf("Error checking collection %s: %v", collName, err))
			continue
		}
		
		if !exists {
			if err := m.createCollection(ctx, collName); err != nil {
				m.log(fmt.Sprintf("Error creating collection %s: %v", collName, err))
				return err
			}
			m.log(fmt.Sprintf("Created collection: %s", collName))
		}
	}
	
	m.log("All required collections ensured")
	return nil
}

// collectionExists checks if a collection exists
func (m *QdrantManager) collectionExists(ctx context.Context, name string) (bool, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(m.endpoint + "/collections/" + name)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK, nil
}

// createCollection creates a new collection
func (m *QdrantManager) createCollection(ctx context.Context, name string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	payload := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     1536,
			"distance": "Cosine",
		},
	}
	
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "PUT", m.endpoint+"/collections/"+name, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("collection creation failed: status %d", resp.StatusCode)
	}
	
	return nil
}

// GetStatus returns current Qdrant status
func (m *QdrantManager) GetStatus(ctx context.Context) (QdrantStatus, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(m.endpoint + "/collections")
	if err != nil {
		return QdrantStatus{}, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	status := QdrantStatus{
		IsHealthy: resp.StatusCode == http.StatusOK,
	}
	
	if collections, ok := result["collections"].([]interface{}); ok {
		for _, c := range collections {
			if m, ok := c.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok {
					status.Collections = append(status.Collections, name)
				}
			}
		}
	}
	
	return status, nil
}

// Migrate handles schema changes between brain versions
func (m *QdrantManager) Migrate(ctx context.Context, version string) error {
	m.log(fmt.Sprintf("Running migrations for version: %s", version))
	// Future: Add version-specific migrations here
	return nil
}

func (m *QdrantManager) log(msg string) {
	select {
	case m.logChannel <- "[QdrantManager] " + msg:
	default:
		fmt.Println("[QdrantManager]", msg)
	}
}
