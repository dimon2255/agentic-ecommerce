package supabase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StorageClient interacts with Supabase Storage for file uploads.
type StorageClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewStorageClient creates a storage client from the Supabase project URL and service role key.
func NewStorageClient(baseURL, apiKey string) *StorageClient {
	return &StorageClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateSignedUploadURL generates a presigned URL for uploading a file to the given bucket.
// The object is stored under a UUID-prefixed path to avoid collisions.
// Returns (uploadURL, publicURL, error).
func (c *StorageClient) CreateSignedUploadURL(bucket, filename, contentType string) (string, string, error) {
	objectPath := fmt.Sprintf("%s/%s", uuid.New().String(), filename)

	// POST /storage/v1/object/upload/sign/{bucket}/{objectPath}
	url := fmt.Sprintf("%s/storage/v1/object/upload/sign/%s/%s", c.baseURL, bucket, objectPath)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("storage error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	// Build full upload URL (the signed URL is a relative path with token)
	uploadURL := fmt.Sprintf("%s/storage/v1%s", c.baseURL, result.URL)
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.baseURL, bucket, objectPath)

	return uploadURL, publicURL, nil
}

// GetPublicURL returns the public URL for a stored object.
func (c *StorageClient) GetPublicURL(bucket, objectPath string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.baseURL, bucket, objectPath)
}
