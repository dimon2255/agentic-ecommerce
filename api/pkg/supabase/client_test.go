package supabase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFrom_Select_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/categories" {
			t.Errorf("expected path /rest/v1/categories, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("select") != "*" {
			t.Errorf("expected select=*, got %s", r.URL.Query().Get("select"))
		}
		if r.Header.Get("apikey") == "" {
			t.Error("expected apikey header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{{"id": "1", "name": "Test"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	var result []map[string]string
	err := client.From("categories").Select("*").Execute(&result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0]["name"] != "Test" {
		t.Errorf("expected name=Test, got %s", result[0]["name"])
	}
}

func TestFrom_Eq_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("slug") != "eq.electronics" {
			t.Errorf("expected slug=eq.electronics, got %s", r.URL.Query().Get("slug"))
		}
		if r.Header.Get("Accept") != "application/vnd.pgrst.object+json" {
			t.Error("expected single object Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "1", "slug": "electronics"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	var result map[string]string
	err := client.From("categories").Select("*").Eq("slug", "electronics").Single().Execute(&result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["slug"] != "electronics" {
		t.Errorf("expected slug=electronics, got %s", result["slug"])
	}
}

func TestFrom_Insert(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Prefer") != "return=representation" {
			t.Error("expected Prefer: return=representation")
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "Electronics" {
			t.Errorf("expected name=Electronics, got %s", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]map[string]string{{"id": "1", "name": "Electronics"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	var result []map[string]string
	err := client.From("categories").Insert(map[string]string{"name": "Electronics"}).Execute(&result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0]["name"] != "Electronics" {
		t.Errorf("expected name=Electronics, got %s", result[0]["name"])
	}
}

func TestFrom_Error_Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	err := client.From("missing").Select("*").Execute(nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/v1/rpc/my_function" {
			t.Errorf("expected path /rest/v1/rpc/my_function, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]int{{"count": 42}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	var result []map[string]int
	err := client.RPC("my_function", map[string]int{"page": 1}, &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0]["count"] != 42 {
		t.Errorf("expected count=42, got %d", result[0]["count"])
	}
}
