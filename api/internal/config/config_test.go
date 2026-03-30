package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Point to empty dir so no config.yaml is found
	t.Setenv("ESHOP_SUPABASE_SERVICE_ROLE_KEY", "test-key")
	t.Setenv("ESHOP_SUPABASE_JWT_SECRET", "test-secret")

	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Supabase.URL != "http://127.0.0.1:54321" {
		t.Errorf("expected default supabase URL, got %s", cfg.Supabase.URL)
	}
	if cfg.Supabase.Timeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", cfg.Supabase.Timeout)
	}
	if len(cfg.CORS.AllowedOrigins) != 2 {
		t.Errorf("expected 2 CORS origins, got %d", len(cfg.CORS.AllowedOrigins))
	}
	if cfg.CORS.MaxAge != 300 {
		t.Errorf("expected max age 300, got %d", cfg.CORS.MaxAge)
	}
	if cfg.Checkout.PaymentCurrency != "usd" {
		t.Errorf("expected usd, got %s", cfg.Checkout.PaymentCurrency)
	}
	if cfg.Checkout.WebhookMaxBodySize != 65536 {
		t.Errorf("expected 65536, got %d", cfg.Checkout.WebhookMaxBodySize)
	}
}

func TestLoad_YAMLOverrides(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
server:
  port: 8080
supabase:
  url: http://custom:54321
  timeout: 30s
cors:
  max_age: 600
checkout:
  payment_currency: eur
`
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yamlContent), 0644)

	t.Setenv("ESHOP_SUPABASE_SERVICE_ROLE_KEY", "test-key")
	t.Setenv("ESHOP_SUPABASE_JWT_SECRET", "test-secret")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Supabase.URL != "http://custom:54321" {
		t.Errorf("expected custom URL, got %s", cfg.Supabase.URL)
	}
	if cfg.Supabase.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cfg.Supabase.Timeout)
	}
	if cfg.CORS.MaxAge != 600 {
		t.Errorf("expected max age 600, got %d", cfg.CORS.MaxAge)
	}
	if cfg.Checkout.PaymentCurrency != "eur" {
		t.Errorf("expected eur, got %s", cfg.Checkout.PaymentCurrency)
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
server:
  port: 8080
`
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yamlContent), 0644)

	t.Setenv("ESHOP_SUPABASE_SERVICE_ROLE_KEY", "test-key")
	t.Setenv("ESHOP_SUPABASE_JWT_SECRET", "test-secret")
	t.Setenv("ESHOP_SERVER_PORT", "7070")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 7070 {
		t.Errorf("expected env override port 7070, got %d", cfg.Server.Port)
	}
}

func TestLoad_MissingServiceRoleKey(t *testing.T) {
	t.Setenv("ESHOP_SUPABASE_JWT_SECRET", "test-secret")

	_, err := Load(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing service role key")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("ESHOP_SUPABASE_SERVICE_ROLE_KEY", "test-key")

	_, err := Load(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing JWT secret")
	}
}
