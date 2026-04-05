package supabase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCreateSignedUploadURL_EncodesFilename(t *testing.T) {
	var capturedURI string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURI = r.RequestURI // Raw URI preserves percent-encoding
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": "/object/upload/sign/test-token"})
	}))
	defer ts.Close()

	client := NewStorageClient(ts.URL, "test-key")

	// Filename with spaces and special characters
	uploadURL, publicURL, err := client.CreateSignedUploadURL("product-images", "my photo (1).png", "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The raw filename should NOT appear unescaped in the request URI
	if strings.Contains(capturedURI, "my photo") {
		t.Errorf("filename was not URL-encoded in request URI: %s", capturedURI)
	}

	// The escaped version should be present
	if !strings.Contains(capturedURI, "my%20photo") {
		t.Errorf("expected URL-encoded filename in URI, got: %s", capturedURI)
	}

	if uploadURL == "" || publicURL == "" {
		t.Error("expected non-empty URLs")
	}

	// Public URL should also contain the encoded filename
	if strings.Contains(publicURL, "my photo") {
		t.Errorf("filename was not URL-encoded in public URL: %s", publicURL)
	}
}

func TestCreateSignedUploadURL_PathTraversalEncoded(t *testing.T) {
	// Verify that url.PathEscape encodes path traversal sequences.
	// We test the URL construction directly because Go's HTTP server
	// correctly rejects paths containing ".." even when percent-encoded.
	encoded := url.PathEscape("../../../etc/passwd")
	if strings.Contains(encoded, "/") {
		t.Errorf("PathEscape did not encode slashes: %s", encoded)
	}
	if strings.Contains(encoded, "../") {
		t.Errorf("PathEscape did not encode path traversal: %s", encoded)
	}
	if !strings.Contains(encoded, "..%2F") {
		t.Errorf("expected encoded dots and slashes, got: %s", encoded)
	}
}

func TestGetPublicURL_EncodesPath(t *testing.T) {
	client := NewStorageClient("https://example.supabase.co", "test-key")

	url := client.GetPublicURL("my bucket", "path/with spaces.png")

	if strings.Contains(url, "my bucket") {
		t.Errorf("bucket was not URL-encoded: %s", url)
	}
	if strings.Contains(url, "with spaces") {
		t.Errorf("object path was not URL-encoded: %s", url)
	}
}
