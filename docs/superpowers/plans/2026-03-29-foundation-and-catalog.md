# Foundation & Catalog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Set up the project infrastructure and build a working, browsable product catalog with category-driven attributes and SKU variants.

**Architecture:** Nuxt 3 (SSR) calls a Go REST API, which communicates with PostgreSQL through Supabase's PostgREST API. Supabase CLI manages local development and migrations.

**Tech Stack:** Nuxt 3, Vue 3, Tailwind CSS, Go 1.22+, chi router, Supabase (PostgREST, Auth, CLI), PostgreSQL, Vitest, Go testing

---

## File Map

### Go API (`api/`)

| File | Responsibility |
|------|---------------|
| `api/cmd/server/main.go` | Entry point, router setup, middleware |
| `api/internal/catalog/models.go` | Request/response types for catalog domain |
| `api/internal/catalog/handler_categories.go` | Category CRUD handlers |
| `api/internal/catalog/handler_categories_test.go` | Category handler tests |
| `api/internal/catalog/handler_attributes.go` | Category attribute + option CRUD handlers |
| `api/internal/catalog/handler_attributes_test.go` | Attribute handler tests |
| `api/internal/catalog/handler_products.go` | Product CRUD handlers |
| `api/internal/catalog/handler_products_test.go` | Product handler tests |
| `api/internal/catalog/handler_skus.go` | SKU CRUD handlers |
| `api/internal/catalog/handler_skus_test.go` | SKU handler tests |
| `api/internal/catalog/handler_custom_fields.go` | Custom field CRUD handlers |
| `api/internal/catalog/handler_custom_fields_test.go` | Custom field handler tests |
| `api/internal/middleware/cors.go` | CORS middleware |
| `api/internal/middleware/logging.go` | Request logging middleware |
| `api/pkg/supabase/client.go` | Supabase REST client wrapper |
| `api/pkg/supabase/client_test.go` | Client wrapper tests |
| `api/pkg/response/json.go` | JSON response helpers |

### Nuxt Frontend (`frontend/`)

| File | Responsibility |
|------|---------------|
| `frontend/nuxt.config.ts` | Nuxt configuration, SSR/SPA route rules |
| `frontend/tailwind.config.ts` | Tailwind CSS configuration |
| `frontend/layouts/default.vue` | Main layout with header navigation |
| `frontend/pages/index.vue` | Homepage — featured categories (SSR) |
| `frontend/pages/catalog/index.vue` | All categories listing (SSR) |
| `frontend/pages/catalog/[slug].vue` | Category page with product grid (SSR) |
| `frontend/pages/product/[slug].vue` | Product detail with SKU selector (SSR) |
| `frontend/components/CategoryCard.vue` | Category card for listings |
| `frontend/components/ProductCard.vue` | Product card for grid display |
| `frontend/components/SkuSelector.vue` | SKU variant selector (attribute dropdowns) |
| `frontend/components/PriceDisplay.vue` | Price display with override logic |
| `frontend/composables/useApi.ts` | Go API client composable |

### Supabase (`supabase/`)

| File | Responsibility |
|------|---------------|
| `supabase/migrations/00001_catalog_schema.sql` | Categories, attributes, options tables |
| `supabase/migrations/00002_products_skus.sql` | Products, SKUs, SKU attribute values, custom fields |
| `supabase/seed.sql` | Sample data for development |

---

## Task 1: Project Scaffolding

**Files:**
- Create: `api/go.mod`, `api/cmd/server/main.go`
- Create: `frontend/` (via npx nuxi init)
- Create: `supabase/config.toml` (via supabase init)
- Create: `.env.example`

- [ ] **Step 1: Initialize Go module**

```bash
mkdir -p api/cmd/server api/internal api/pkg
cd api
go mod init github.com/dimon2255/agentic-ecommerce/api
```

- [ ] **Step 2: Install Go dependencies**

```bash
cd api
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
```

- [ ] **Step 3: Create Go server entry point**

Create `api/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("API server listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

- [ ] **Step 4: Verify Go server compiles and runs**

```bash
cd api
go run cmd/server/main.go
```

Expected: `API server listening on :8080`

Test health endpoint in another terminal:
```bash
curl http://localhost:8080/health
```
Expected: `{"status":"ok"}`

- [ ] **Step 5: Initialize Nuxt 3 project**

```bash
npx nuxi@latest init frontend
cd frontend
npm install
```

- [ ] **Step 6: Install Nuxt dependencies**

```bash
cd frontend
npm install -D @nuxtjs/tailwindcss
npm install -D @nuxtjs/supabase
```

- [ ] **Step 7: Configure Nuxt**

Create `frontend/nuxt.config.ts`:

```ts
export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/supabase'],

  supabase: {
    redirect: false,
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
    },
  },

  routeRules: {
    '/': { ssr: true },
    '/catalog/**': { ssr: true },
    '/product/**': { ssr: true },
    '/cart': { ssr: false },
    '/checkout': { ssr: false },
    '/account/**': { ssr: false },
  },

  compatibilityDate: '2025-01-01',
})
```

- [ ] **Step 8: Initialize Tailwind**

Create `frontend/tailwind.config.ts`:

```ts
import type { Config } from 'tailwindcss'

export default {
  content: [
    './components/**/*.{vue,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './composables/**/*.ts',
    './app.vue',
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
      },
    },
  },
  plugins: [],
} satisfies Config
```

- [ ] **Step 9: Initialize Supabase**

```bash
supabase init
```

This creates `supabase/config.toml` in the project root.

- [ ] **Step 10: Start Supabase locally**

```bash
supabase start
```

Expected output includes:
```
API URL: http://127.0.0.1:54321
anon key: eyJ...
service_role key: eyJ...
DB URL: postgresql://postgres:postgres@127.0.0.1:54322/postgres
```

Save these values.

- [ ] **Step 11: Create .env files**

Create `.env.example`:

```env
# Supabase
SUPABASE_URL=http://127.0.0.1:54321
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Go API
API_PORT=8080

# Nuxt
NUXT_PUBLIC_API_BASE=http://localhost:8080
NUXT_PUBLIC_SUPABASE_URL=http://127.0.0.1:54321
NUXT_PUBLIC_SUPABASE_KEY=your-anon-key
```

Copy to `.env` and fill in the actual keys from `supabase start` output.

- [ ] **Step 12: Commit**

```bash
git add -A
git commit -m "feat: initialize project scaffolding (Nuxt 3, Go, Supabase)"
```

---

## Task 2: Database Migrations — Catalog Schema

**Files:**
- Create: `supabase/migrations/00001_catalog_schema.sql`

- [ ] **Step 1: Create catalog schema migration**

Create `supabase/migrations/00001_catalog_schema.sql`:

```sql
-- Categories (hierarchical via parent_id)
create table categories (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  slug text unique not null,
  parent_id uuid references categories(id) on delete set null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_categories_slug on categories(slug);
create index idx_categories_parent_id on categories(parent_id);

-- Category attributes (defines what attributes a category's products have)
create type attribute_type as enum ('text', 'number', 'enum');

create table category_attributes (
  id uuid primary key default gen_random_uuid(),
  category_id uuid not null references categories(id) on delete cascade,
  name text not null,
  type attribute_type not null default 'text',
  required boolean not null default false,
  sort_order integer not null default 0
);

create index idx_category_attributes_category on category_attributes(category_id);

-- Attribute options (enum values for enum-type attributes)
create table attribute_options (
  id uuid primary key default gen_random_uuid(),
  category_attribute_id uuid not null references category_attributes(id) on delete cascade,
  value text not null,
  sort_order integer not null default 0
);

create index idx_attribute_options_attr on attribute_options(category_attribute_id);

-- Enable RLS
alter table categories enable row level security;
alter table category_attributes enable row level security;
alter table attribute_options enable row level security;

-- Public read access for catalog
create policy "categories_public_read" on categories
  for select using (true);

create policy "category_attributes_public_read" on category_attributes
  for select using (true);

create policy "attribute_options_public_read" on attribute_options
  for select using (true);

-- Service role has full access (Go API uses service_role key)
create policy "categories_service_all" on categories
  for all using (auth.role() = 'service_role');

create policy "category_attributes_service_all" on category_attributes
  for all using (auth.role() = 'service_role');

create policy "attribute_options_service_all" on attribute_options
  for all using (auth.role() = 'service_role');

-- Updated_at trigger
create or replace function update_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger categories_updated_at
  before update on categories
  for each row execute function update_updated_at();
```

- [ ] **Step 2: Apply migration**

```bash
supabase db reset
```

Expected: Migration applies without errors.

- [ ] **Step 3: Verify tables exist**

```bash
supabase db lint
```

Or check via Supabase Studio at `http://127.0.0.1:54323`.

- [ ] **Step 4: Commit**

```bash
git add supabase/migrations/00001_catalog_schema.sql
git commit -m "feat: add catalog schema migration (categories, attributes, options)"
```

---

## Task 3: Database Migrations — Products & SKUs

**Files:**
- Create: `supabase/migrations/00002_products_skus.sql`

- [ ] **Step 1: Create products and SKUs migration**

Create `supabase/migrations/00002_products_skus.sql`:

```sql
-- Products
create type product_status as enum ('draft', 'active', 'archived');

create table products (
  id uuid primary key default gen_random_uuid(),
  category_id uuid not null references categories(id) on delete restrict,
  name text not null,
  slug text unique not null,
  description text,
  base_price numeric(10,2) not null check (base_price >= 0),
  status product_status not null default 'draft',
  images text[] default '{}',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_products_slug on products(slug);
create index idx_products_category on products(category_id);
create index idx_products_status on products(status);

-- SKUs
create type sku_status as enum ('active', 'inactive');

create table skus (
  id uuid primary key default gen_random_uuid(),
  product_id uuid not null references products(id) on delete cascade,
  sku_code text unique not null,
  price_override numeric(10,2) check (price_override >= 0),
  status sku_status not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_skus_product on skus(product_id);
create index idx_skus_code on skus(sku_code);

-- SKU attribute values
create table sku_attribute_values (
  id uuid primary key default gen_random_uuid(),
  sku_id uuid not null references skus(id) on delete cascade,
  category_attribute_id uuid not null references category_attributes(id) on delete cascade,
  value text not null
);

create index idx_sku_attr_values_sku on sku_attribute_values(sku_id);

-- Custom fields (reporting metadata only)
create type entity_type as enum ('product', 'sku');

create table custom_fields (
  id uuid primary key default gen_random_uuid(),
  entity_type entity_type not null,
  entity_id uuid not null,
  key text not null,
  value text not null
);

create index idx_custom_fields_entity on custom_fields(entity_type, entity_id);
create index idx_custom_fields_key on custom_fields(key);

-- Page views (for admin report)
create table page_views (
  id uuid primary key default gen_random_uuid(),
  product_id uuid not null references products(id) on delete cascade,
  session_id text,
  viewed_at timestamptz not null default now()
);

create index idx_page_views_product on page_views(product_id);

-- RLS policies
alter table products enable row level security;
alter table skus enable row level security;
alter table sku_attribute_values enable row level security;
alter table custom_fields enable row level security;
alter table page_views enable row level security;

-- Public read
create policy "products_public_read" on products
  for select using (status = 'active');

create policy "skus_public_read" on skus
  for select using (status = 'active');

create policy "sku_attr_values_public_read" on sku_attribute_values
  for select using (true);

create policy "custom_fields_public_read" on custom_fields
  for select using (true);

create policy "page_views_public_insert" on page_views
  for insert with check (true);

-- Service role full access
create policy "products_service_all" on products
  for all using (auth.role() = 'service_role');

create policy "skus_service_all" on skus
  for all using (auth.role() = 'service_role');

create policy "sku_attr_values_service_all" on sku_attribute_values
  for all using (auth.role() = 'service_role');

create policy "custom_fields_service_all" on custom_fields
  for all using (auth.role() = 'service_role');

create policy "page_views_service_all" on page_views
  for all using (auth.role() = 'service_role');

-- Updated_at triggers
create trigger products_updated_at
  before update on products
  for each row execute function update_updated_at();

create trigger skus_updated_at
  before update on skus
  for each row execute function update_updated_at();
```

- [ ] **Step 2: Apply migration**

```bash
supabase db reset
```

Expected: Both migrations apply without errors.

- [ ] **Step 3: Commit**

```bash
git add supabase/migrations/00002_products_skus.sql
git commit -m "feat: add products, SKUs, custom fields, page views schema"
```

---

## Task 4: Go API — Supabase Client Wrapper

**Files:**
- Create: `api/pkg/supabase/client.go`
- Create: `api/pkg/supabase/client_test.go`
- Create: `api/pkg/response/json.go`

- [ ] **Step 1: Create JSON response helpers**

Create `api/pkg/response/json.go`:

```go
package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
```

- [ ] **Step 2: Create Supabase client wrapper**

Create `api/pkg/supabase/client.go`:

```go
package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueryParams builds PostgREST query parameters.
type QueryParams map[string]string

// From performs a GET request on a table with optional PostgREST filters.
// Example: client.From("categories").Select("*").Eq("slug", "electronics").Execute()
type QueryBuilder struct {
	client  *Client
	table   string
	params  url.Values
	method  string
	body    any
	headers map[string]string
	single  bool
}

func (c *Client) From(table string) *QueryBuilder {
	return &QueryBuilder{
		client:  c,
		table:   table,
		params:  url.Values{},
		method:  "GET",
		headers: map[string]string{},
	}
}

func (q *QueryBuilder) Select(columns string) *QueryBuilder {
	q.params.Set("select", columns)
	return q
}

func (q *QueryBuilder) Eq(column, value string) *QueryBuilder {
	q.params.Set(column, "eq."+value)
	return q
}

func (q *QueryBuilder) Is(column, value string) *QueryBuilder {
	q.params.Set(column, "is."+value)
	return q
}

func (q *QueryBuilder) In(column string, values []string) *QueryBuilder {
	joined := ""
	for i, v := range values {
		if i > 0 {
			joined += ","
		}
		joined += `"` + v + `"`
	}
	q.params.Set(column, "in.("+joined+")")
	return q
}

func (q *QueryBuilder) Order(column, direction string) *QueryBuilder {
	q.params.Set("order", column+"."+direction)
	return q
}

func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.params.Set("limit", fmt.Sprintf("%d", n))
	return q
}

func (q *QueryBuilder) Offset(n int) *QueryBuilder {
	q.params.Set("offset", fmt.Sprintf("%d", n))
	return q
}

func (q *QueryBuilder) Single() *QueryBuilder {
	q.single = true
	q.headers["Accept"] = "application/vnd.pgrst.object+json"
	return q
}

func (q *QueryBuilder) Insert(data any) *QueryBuilder {
	q.method = "POST"
	q.body = data
	q.headers["Prefer"] = "return=representation"
	return q
}

func (q *QueryBuilder) Update(data any) *QueryBuilder {
	q.method = "PATCH"
	q.body = data
	q.headers["Prefer"] = "return=representation"
	return q
}

func (q *QueryBuilder) Delete() *QueryBuilder {
	q.method = "DELETE"
	q.headers["Prefer"] = "return=representation"
	return q
}

func (q *QueryBuilder) Execute(result any) error {
	reqURL := fmt.Sprintf("%s/rest/v1/%s", q.client.baseURL, q.table)
	if len(q.params) > 0 {
		reqURL += "?" + q.params.Encode()
	}

	var bodyReader io.Reader
	if q.body != nil {
		jsonBody, err := json.Marshal(q.body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(q.method, reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("apikey", q.client.apiKey)
	req.Header.Set("Authorization", "Bearer "+q.client.apiKey)
	req.Header.Set("Content-Type", "application/json")

	for k, v := range q.headers {
		req.Header.Set(k, v)
	}

	resp, err := q.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

// RPC calls a PostgreSQL function via Supabase RPC endpoint.
func (c *Client) RPC(functionName string, params any, result any) error {
	reqURL := fmt.Sprintf("%s/rest/v1/rpc/%s", c.baseURL, functionName)

	var bodyReader io.Reader
	if params != nil {
		jsonBody, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("POST", reqURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase rpc error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}
```

- [ ] **Step 3: Write Supabase client tests**

Create `api/pkg/supabase/client_test.go`:

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd api
go test ./pkg/supabase/ -v
```

Expected: All 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/pkg/supabase/ api/pkg/response/
git commit -m "feat: add Supabase REST client wrapper and JSON response helpers"
```

---

## Task 5: Go API — Catalog Models

**Files:**
- Create: `api/internal/catalog/models.go`

- [ ] **Step 1: Define catalog domain types**

Create `api/internal/catalog/models.go`:

```go
package catalog

import "time"

// --- Categories ---

type Category struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	ParentID  *string    `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ParentID *string `json:"parent_id,omitempty"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

// --- Category Attributes ---

type CategoryAttribute struct {
	ID         string            `json:"id"`
	CategoryID string            `json:"category_id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Required   bool              `json:"required"`
	SortOrder  int               `json:"sort_order"`
	Options    []AttributeOption `json:"options,omitempty"`
}

type CreateAttributeRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	SortOrder int    `json:"sort_order"`
}

type AttributeOption struct {
	ID                  string `json:"id"`
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
	SortOrder           int    `json:"sort_order"`
}

type CreateAttributeOptionRequest struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// --- Products ---

type Product struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	BasePrice   float64   `json:"base_price"`
	Status      string    `json:"status"`
	Images      []string  `json:"images"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	CategoryID  string   `json:"category_id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description *string  `json:"description,omitempty"`
	BasePrice   float64  `json:"base_price"`
	Status      string   `json:"status"`
	Images      []string `json:"images,omitempty"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Slug        *string  `json:"slug,omitempty"`
	Description *string  `json:"description,omitempty"`
	BasePrice   *float64 `json:"base_price,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Images      []string `json:"images,omitempty"`
}

// --- SKUs ---

type SKU struct {
	ID              string              `json:"id"`
	ProductID       string              `json:"product_id"`
	SKUCode         string              `json:"sku_code"`
	PriceOverride   *float64            `json:"price_override"`
	Status          string              `json:"status"`
	AttributeValues []SKUAttributeValue `json:"attribute_values,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type CreateSKURequest struct {
	SKUCode         string                        `json:"sku_code"`
	PriceOverride   *float64                      `json:"price_override,omitempty"`
	Status          string                        `json:"status"`
	AttributeValues []CreateSKUAttributeValueReq  `json:"attribute_values"`
}

type CreateSKUAttributeValueReq struct {
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
}

type SKUAttributeValue struct {
	ID                  string `json:"id"`
	SKUID               string `json:"sku_id"`
	CategoryAttributeID string `json:"category_attribute_id"`
	Value               string `json:"value"`
}

// --- Custom Fields ---

type CustomField struct {
	ID         string `json:"id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}

type CreateCustomFieldRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
```

- [ ] **Step 2: Commit**

```bash
git add api/internal/catalog/models.go
git commit -m "feat: add catalog domain models and request types"
```

---

## Task 6: Go API — Category Handlers

**Files:**
- Create: `api/internal/catalog/handler_categories.go`
- Create: `api/internal/catalog/handler_categories_test.go`

- [ ] **Step 1: Write category handler tests**

Create `api/internal/catalog/handler_categories_test.go`:

```go
package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestCategoryHandler(supabaseHandler http.HandlerFunc) (*CategoryHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewCategoryHandler(client)
	return handler, server
}

func TestListCategories(t *testing.T) {
	handler, server := setupTestCategoryHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Category{
			{ID: "1", Name: "Electronics", Slug: "electronics"},
			{ID: "2", Name: "Clothing", Slug: "clothing"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/categories", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []Category
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(result))
	}
}

func TestGetCategoryBySlug(t *testing.T) {
	handler, server := setupTestCategoryHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Category{ID: "1", Name: "Electronics", Slug: "electronics"})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/categories/electronics", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "electronics")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetBySlug(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result Category
	json.NewDecoder(w.Body).Decode(&result)
	if result.Slug != "electronics" {
		t.Errorf("expected slug=electronics, got %s", result.Slug)
	}
}

func TestCreateCategory(t *testing.T) {
	handler, server := setupTestCategoryHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]Category{{ID: "1", Name: "Electronics", Slug: "electronics"}})
	})
	defer server.Close()

	body := `{"name":"Electronics","slug":"electronics"}`
	req := httptest.NewRequest("POST", "/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd api
go test ./internal/catalog/ -v -run TestListCategories
```

Expected: FAIL — `NewCategoryHandler` not defined.

- [ ] **Step 3: Implement category handlers**

Create `api/internal/catalog/handler_categories.go`:

```go
package catalog

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CategoryHandler struct {
	db *supabase.Client
}

func NewCategoryHandler(db *supabase.Client) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{slug}", h.GetBySlug)
	r.Patch("/{slug}", h.Update)
	r.Delete("/{slug}", h.DeleteBySlug)
	return r
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	parentID := r.URL.Query().Get("parent_id")

	query := h.db.From("categories").Select("*").Order("name", "asc")
	if parentID == "null" {
		query = query.Is("parent_id", "null")
	} else if parentID != "" {
		query = query.Eq("parent_id", parentID)
	}

	var categories []Category
	if err := query.Execute(&categories); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	if categories == nil {
		categories = []Category{}
	}
	response.JSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var category Category
	err := h.db.From("categories").Select("*").Eq("slug", slug).Single().Execute(&category)
	if err != nil {
		response.Error(w, http.StatusNotFound, "category not found")
		return
	}

	response.JSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		response.Error(w, http.StatusBadRequest, "name and slug are required")
		return
	}

	var created []Category
	if err := h.db.From("categories").Insert(req).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create category")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var updated []Category
	err := h.db.From("categories").Eq("slug", slug).Update(req).Execute(&updated)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update category")
		return
	}

	if len(updated) == 0 {
		response.Error(w, http.StatusNotFound, "category not found")
		return
	}

	response.JSON(w, http.StatusOK, updated[0])
}

func (h *CategoryHandler) DeleteBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	err := h.db.From("categories").Eq("slug", slug).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Fix test import and run tests**

Add `"context"` import to the test file, then run:

```bash
cd api
go test ./internal/catalog/ -v
```

Expected: All 3 category tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/catalog/handler_categories.go api/internal/catalog/handler_categories_test.go
git commit -m "feat: add category CRUD handlers with tests"
```

---

## Task 7: Go API — Category Attributes & Options Handlers

**Files:**
- Create: `api/internal/catalog/handler_attributes.go`
- Create: `api/internal/catalog/handler_attributes_test.go`

- [ ] **Step 1: Write attribute handler tests**

Create `api/internal/catalog/handler_attributes_test.go`:

```go
package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestAttributeHandler(supabaseHandler http.HandlerFunc) (*AttributeHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewAttributeHandler(client)
	return handler, server
}

func TestListAttributes(t *testing.T) {
	handler, server := setupTestAttributeHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CategoryAttribute{
			{ID: "1", CategoryID: "cat-1", Name: "Size", Type: "enum"},
			{ID: "2", CategoryID: "cat-1", Name: "Color", Type: "enum"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/categories/cat-1/attributes", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("categoryId", "cat-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []CategoryAttribute
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(result))
	}
}

func TestCreateAttribute(t *testing.T) {
	handler, server := setupTestAttributeHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]CategoryAttribute{{ID: "1", CategoryID: "cat-1", Name: "Size", Type: "enum"}})
	})
	defer server.Close()

	body := `{"name":"Size","type":"enum","required":true,"sort_order":0}`
	req := httptest.NewRequest("POST", "/categories/cat-1/attributes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("categoryId", "cat-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd api
go test ./internal/catalog/ -v -run TestListAttributes
```

Expected: FAIL — `NewAttributeHandler` not defined.

- [ ] **Step 3: Implement attribute handlers**

Create `api/internal/catalog/handler_attributes.go`:

```go
package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type AttributeHandler struct {
	db *supabase.Client
}

func NewAttributeHandler(db *supabase.Client) *AttributeHandler {
	return &AttributeHandler{db: db}
}

func (h *AttributeHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{attrId}", h.Delete)

	// Attribute options
	r.Get("/{attrId}/options", h.ListOptions)
	r.Post("/{attrId}/options", h.CreateOption)
	r.Delete("/{attrId}/options/{optionId}", h.DeleteOption)
	return r
}

func (h *AttributeHandler) List(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")

	var attrs []CategoryAttribute
	err := h.db.From("category_attributes").
		Select("*").
		Eq("category_id", categoryID).
		Order("sort_order", "asc").
		Execute(&attrs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch attributes")
		return
	}

	// Fetch options for each attribute
	for i := range attrs {
		var options []AttributeOption
		h.db.From("attribute_options").
			Select("*").
			Eq("category_attribute_id", attrs[i].ID).
			Order("sort_order", "asc").
			Execute(&options)
		if options == nil {
			options = []AttributeOption{}
		}
		attrs[i].Options = options
	}

	if attrs == nil {
		attrs = []CategoryAttribute{}
	}
	response.JSON(w, http.StatusOK, attrs)
}

func (h *AttributeHandler) Create(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")

	var req CreateAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Type == "" {
		response.Error(w, http.StatusBadRequest, "name and type are required")
		return
	}

	insertData := map[string]any{
		"category_id": categoryID,
		"name":        req.Name,
		"type":        req.Type,
		"required":    req.Required,
		"sort_order":  req.SortOrder,
	}

	var created []CategoryAttribute
	if err := h.db.From("category_attributes").Insert(insertData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create attribute")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *AttributeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")

	err := h.db.From("category_attributes").Eq("id", attrID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete attribute")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AttributeHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")

	var options []AttributeOption
	err := h.db.From("attribute_options").
		Select("*").
		Eq("category_attribute_id", attrID).
		Order("sort_order", "asc").
		Execute(&options)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch options")
		return
	}

	if options == nil {
		options = []AttributeOption{}
	}
	response.JSON(w, http.StatusOK, options)
}

func (h *AttributeHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	attrID := chi.URLParam(r, "attrId")

	var req CreateAttributeOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	insertData := map[string]any{
		"category_attribute_id": attrID,
		"value":                 req.Value,
		"sort_order":            req.SortOrder,
	}

	var created []AttributeOption
	if err := h.db.From("attribute_options").Insert(insertData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create option")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *AttributeHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "optionId")

	err := h.db.From("attribute_options").Eq("id", optionID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete option")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run all tests**

```bash
cd api
go test ./internal/catalog/ -v
```

Expected: All category and attribute tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/catalog/handler_attributes.go api/internal/catalog/handler_attributes_test.go
git commit -m "feat: add category attribute and option CRUD handlers with tests"
```

---

## Task 8: Go API — Product Handlers

**Files:**
- Create: `api/internal/catalog/handler_products.go`
- Create: `api/internal/catalog/handler_products_test.go`

- [ ] **Step 1: Write product handler tests**

Create `api/internal/catalog/handler_products_test.go`:

```go
package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestProductHandler(supabaseHandler http.HandlerFunc) (*ProductHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewProductHandler(client)
	return handler, server
}

func TestListProducts(t *testing.T) {
	handler, server := setupTestProductHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Product{
			{ID: "1", Name: "T-Shirt", Slug: "t-shirt", BasePrice: 29.99, Status: "active"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/products?category_id=cat-1", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []Product
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 product, got %d", len(result))
	}
}

func TestGetProductBySlug(t *testing.T) {
	handler, server := setupTestProductHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Product{ID: "1", Name: "T-Shirt", Slug: "t-shirt", BasePrice: 29.99})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/products/t-shirt", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "t-shirt")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetBySlug(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result Product
	json.NewDecoder(w.Body).Decode(&result)
	if result.Slug != "t-shirt" {
		t.Errorf("expected slug=t-shirt, got %s", result.Slug)
	}
}

func TestCreateProduct(t *testing.T) {
	handler, server := setupTestProductHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]Product{{ID: "1", Name: "T-Shirt", Slug: "t-shirt", BasePrice: 29.99}})
	})
	defer server.Close()

	body := `{"category_id":"cat-1","name":"T-Shirt","slug":"t-shirt","base_price":29.99,"status":"draft"}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd api
go test ./internal/catalog/ -v -run TestListProducts
```

Expected: FAIL — `NewProductHandler` not defined.

- [ ] **Step 3: Implement product handlers**

Create `api/internal/catalog/handler_products.go`:

```go
package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type ProductHandler struct {
	db *supabase.Client
}

func NewProductHandler(db *supabase.Client) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{slug}", h.GetBySlug)
	r.Patch("/{slug}", h.Update)
	r.Delete("/{slug}", h.DeleteBySlug)
	return r
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("category_id")

	query := h.db.From("products").Select("*").Order("created_at", "desc")
	if categoryID != "" {
		query = query.Eq("category_id", categoryID)
	}

	var products []Product
	if err := query.Execute(&products); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}

	if products == nil {
		products = []Product{}
	}
	response.JSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var product Product
	err := h.db.From("products").Select("*").Eq("slug", slug).Single().Execute(&product)
	if err != nil {
		response.Error(w, http.StatusNotFound, "product not found")
		return
	}

	response.JSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" || req.CategoryID == "" {
		response.Error(w, http.StatusBadRequest, "name, slug, and category_id are required")
		return
	}

	if req.Status == "" {
		req.Status = "draft"
	}

	var created []Product
	if err := h.db.From("products").Insert(req).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create product")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var updated []Product
	err := h.db.From("products").Eq("slug", slug).Update(req).Execute(&updated)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update product")
		return
	}

	if len(updated) == 0 {
		response.Error(w, http.StatusNotFound, "product not found")
		return
	}

	response.JSON(w, http.StatusOK, updated[0])
}

func (h *ProductHandler) DeleteBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	err := h.db.From("products").Eq("slug", slug).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run tests**

```bash
cd api
go test ./internal/catalog/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/catalog/handler_products.go api/internal/catalog/handler_products_test.go
git commit -m "feat: add product CRUD handlers with tests"
```

---

## Task 9: Go API — SKU Handlers

**Files:**
- Create: `api/internal/catalog/handler_skus.go`
- Create: `api/internal/catalog/handler_skus_test.go`

- [ ] **Step 1: Write SKU handler tests**

Create `api/internal/catalog/handler_skus_test.go`:

```go
package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestSKUHandler(supabaseHandler http.HandlerFunc) (*SKUHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewSKUHandler(client)
	return handler, server
}

func TestListSKUs(t *testing.T) {
	callCount := 0
	handler, server := setupTestSKUHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			// First call: list SKUs
			json.NewEncoder(w).Encode([]SKU{
				{ID: "sku-1", ProductID: "prod-1", SKUCode: "TSHIRT-BLU-M", Status: "active"},
			})
		} else {
			// Second call: fetch attribute values
			json.NewEncoder(w).Encode([]SKUAttributeValue{
				{ID: "av-1", SKUID: "sku-1", CategoryAttributeID: "attr-1", Value: "Blue"},
			})
		}
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/products/prod-1/skus", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("productId", "prod-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []SKU
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 SKU, got %d", len(result))
	}
	if result[0].SKUCode != "TSHIRT-BLU-M" {
		t.Errorf("expected SKU code TSHIRT-BLU-M, got %s", result[0].SKUCode)
	}
}

func TestCreateSKU(t *testing.T) {
	callCount := 0
	handler, server := setupTestSKUHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			// Insert SKU
			w.WriteHeader(201)
			json.NewEncoder(w).Encode([]SKU{{ID: "sku-1", ProductID: "prod-1", SKUCode: "TSHIRT-BLU-M"}})
		} else {
			// Insert attribute values
			w.WriteHeader(201)
			json.NewEncoder(w).Encode([]SKUAttributeValue{{ID: "av-1", SKUID: "sku-1"}})
		}
	})
	defer server.Close()

	body := `{"sku_code":"TSHIRT-BLU-M","status":"active","attribute_values":[{"category_attribute_id":"attr-1","value":"Blue"}]}`
	req := httptest.NewRequest("POST", "/products/prod-1/skus", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("productId", "prod-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd api
go test ./internal/catalog/ -v -run TestListSKUs
```

Expected: FAIL — `NewSKUHandler` not defined.

- [ ] **Step 3: Implement SKU handlers**

Create `api/internal/catalog/handler_skus.go`:

```go
package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type SKUHandler struct {
	db *supabase.Client
}

func NewSKUHandler(db *supabase.Client) *SKUHandler {
	return &SKUHandler{db: db}
}

func (h *SKUHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{skuId}", h.Delete)
	return r
}

func (h *SKUHandler) List(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")

	var skus []SKU
	err := h.db.From("skus").
		Select("*").
		Eq("product_id", productID).
		Order("created_at", "asc").
		Execute(&skus)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch SKUs")
		return
	}

	// Fetch attribute values for each SKU
	for i := range skus {
		var attrValues []SKUAttributeValue
		h.db.From("sku_attribute_values").
			Select("*").
			Eq("sku_id", skus[i].ID).
			Execute(&attrValues)
		if attrValues == nil {
			attrValues = []SKUAttributeValue{}
		}
		skus[i].AttributeValues = attrValues
	}

	if skus == nil {
		skus = []SKU{}
	}
	response.JSON(w, http.StatusOK, skus)
}

func (h *SKUHandler) Create(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "productId")

	var req CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SKUCode == "" {
		response.Error(w, http.StatusBadRequest, "sku_code is required")
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	// Insert SKU
	skuData := map[string]any{
		"product_id":     productID,
		"sku_code":       req.SKUCode,
		"price_override": req.PriceOverride,
		"status":         req.Status,
	}

	var created []SKU
	if err := h.db.From("skus").Insert(skuData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create SKU")
		return
	}

	sku := created[0]

	// Insert attribute values
	for _, av := range req.AttributeValues {
		avData := map[string]any{
			"sku_id":                sku.ID,
			"category_attribute_id": av.CategoryAttributeID,
			"value":                 av.Value,
		}
		h.db.From("sku_attribute_values").Insert(avData).Execute(nil)
	}

	// Fetch the attribute values back for the response
	var attrValues []SKUAttributeValue
	h.db.From("sku_attribute_values").Select("*").Eq("sku_id", sku.ID).Execute(&attrValues)
	sku.AttributeValues = attrValues

	response.JSON(w, http.StatusCreated, sku)
}

func (h *SKUHandler) Delete(w http.ResponseWriter, r *http.Request) {
	skuID := chi.URLParam(r, "skuId")

	err := h.db.From("skus").Eq("id", skuID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete SKU")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run all tests**

```bash
cd api
go test ./internal/catalog/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add api/internal/catalog/handler_skus.go api/internal/catalog/handler_skus_test.go
git commit -m "feat: add SKU CRUD handlers with attribute values and tests"
```

---

## Task 10: Go API — Custom Fields Handlers

**Files:**
- Create: `api/internal/catalog/handler_custom_fields.go`
- Create: `api/internal/catalog/handler_custom_fields_test.go`

- [ ] **Step 1: Write custom field handler tests**

Create `api/internal/catalog/handler_custom_fields_test.go`:

```go
package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	supa "github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func setupTestCustomFieldHandler(supabaseHandler http.HandlerFunc) (*CustomFieldHandler, *httptest.Server) {
	server := httptest.NewServer(supabaseHandler)
	client := supa.NewClient(server.URL, "test-key")
	handler := NewCustomFieldHandler(client)
	return handler, server
}

func TestListCustomFields(t *testing.T) {
	handler, server := setupTestCustomFieldHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]CustomField{
			{ID: "1", EntityType: "product", EntityID: "prod-1", Key: "supplier", Value: "Acme"},
		})
	})
	defer server.Close()

	req := httptest.NewRequest("GET", "/custom-fields?entity_type=product&entity_id=prod-1", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []CustomField
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 field, got %d", len(result))
	}
}

func TestCreateCustomField(t *testing.T) {
	handler, server := setupTestCustomFieldHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode([]CustomField{{ID: "1", EntityType: "product", EntityID: "prod-1", Key: "supplier", Value: "Acme"}})
	})
	defer server.Close()

	body := `{"key":"supplier","value":"Acme"}`
	req := httptest.NewRequest("POST", "/custom-fields?entity_type=product&entity_id=prod-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd api
go test ./internal/catalog/ -v -run TestListCustomFields
```

Expected: FAIL — `NewCustomFieldHandler` not defined.

- [ ] **Step 3: Implement custom field handlers**

Create `api/internal/catalog/handler_custom_fields.go`:

```go
package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

type CustomFieldHandler struct {
	db *supabase.Client
}

func NewCustomFieldHandler(db *supabase.Client) *CustomFieldHandler {
	return &CustomFieldHandler{db: db}
}

func (h *CustomFieldHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{fieldId}", h.Delete)
	return r
}

func (h *CustomFieldHandler) List(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	if entityType == "" || entityID == "" {
		response.Error(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	var fields []CustomField
	err := h.db.From("custom_fields").
		Select("*").
		Eq("entity_type", entityType).
		Eq("entity_id", entityID).
		Execute(&fields)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch custom fields")
		return
	}

	if fields == nil {
		fields = []CustomField{}
	}
	response.JSON(w, http.StatusOK, fields)
}

func (h *CustomFieldHandler) Create(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")

	if entityType == "" || entityID == "" {
		response.Error(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	var req CreateCustomFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Key == "" || req.Value == "" {
		response.Error(w, http.StatusBadRequest, "key and value are required")
		return
	}

	insertData := map[string]any{
		"entity_type": entityType,
		"entity_id":   entityID,
		"key":         req.Key,
		"value":       req.Value,
	}

	var created []CustomField
	if err := h.db.From("custom_fields").Insert(insertData).Execute(&created); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create custom field")
		return
	}

	response.JSON(w, http.StatusCreated, created[0])
}

func (h *CustomFieldHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	err := h.db.From("custom_fields").Eq("id", fieldID).Delete().Execute(nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete custom field")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 4: Run all tests**

```bash
cd api
go test ./... -v
```

Expected: All tests PASS across all packages.

- [ ] **Step 5: Commit**

```bash
git add api/internal/catalog/handler_custom_fields.go api/internal/catalog/handler_custom_fields_test.go
git commit -m "feat: add custom field CRUD handlers with tests"
```

---

## Task 11: Go API — Wire Up Router

**Files:**
- Modify: `api/cmd/server/main.go`

- [ ] **Step 1: Wire all handlers into the router**

Update `api/cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "http://127.0.0.1:54321"
	}
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		log.Fatal("SUPABASE_SERVICE_ROLE_KEY is required")
	}

	db := supabase.NewClient(supabaseURL, supabaseKey)

	categoryHandler := catalog.NewCategoryHandler(db)
	attributeHandler := catalog.NewAttributeHandler(db)
	productHandler := catalog.NewProductHandler(db)
	skuHandler := catalog.NewSKUHandler(db)
	customFieldHandler := catalog.NewCustomFieldHandler(db)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/categories", categoryHandler.Routes())
		r.Route("/categories/{categoryId}/attributes", func(r chi.Router) {
			r.Mount("/", attributeHandler.Routes())
		})
		r.Mount("/products", productHandler.Routes())
		r.Route("/products/{productId}/skus", func(r chi.Router) {
			r.Mount("/", skuHandler.Routes())
		})
		r.Mount("/custom-fields", customFieldHandler.Routes())
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("API server listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd api
go build ./cmd/server/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add api/cmd/server/main.go
git commit -m "feat: wire catalog handlers into API router"
```

---

## Task 12: Seed Data

**Files:**
- Create: `supabase/seed.sql`

- [ ] **Step 1: Create seed data**

Create `supabase/seed.sql`:

```sql
-- Categories
insert into categories (id, name, slug, parent_id) values
  ('a1000000-0000-0000-0000-000000000001', 'Electronics', 'electronics', null),
  ('a1000000-0000-0000-0000-000000000002', 'Clothing', 'clothing', null),
  ('a1000000-0000-0000-0000-000000000003', 'Laptops', 'laptops', 'a1000000-0000-0000-0000-000000000001'),
  ('a1000000-0000-0000-0000-000000000004', 'T-Shirts', 't-shirts', 'a1000000-0000-0000-0000-000000000002');

-- Category Attributes
-- Laptops: RAM, Storage
insert into category_attributes (id, category_id, name, type, required, sort_order) values
  ('b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000003', 'RAM', 'enum', true, 0),
  ('b1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000003', 'Storage', 'enum', true, 1);

-- T-Shirts: Size, Color
insert into category_attributes (id, category_id, name, type, required, sort_order) values
  ('b1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000004', 'Size', 'enum', true, 0),
  ('b1000000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000004', 'Color', 'enum', true, 1);

-- Attribute Options
-- RAM: 8GB, 16GB, 32GB
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000001', '8GB', 0),
  ('b1000000-0000-0000-0000-000000000001', '16GB', 1),
  ('b1000000-0000-0000-0000-000000000001', '32GB', 2);

-- Storage: 256GB, 512GB, 1TB
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000002', '256GB', 0),
  ('b1000000-0000-0000-0000-000000000002', '512GB', 1),
  ('b1000000-0000-0000-0000-000000000002', '1TB', 2);

-- Size: S, M, L, XL
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000003', 'S', 0),
  ('b1000000-0000-0000-0000-000000000003', 'M', 1),
  ('b1000000-0000-0000-0000-000000000003', 'L', 2),
  ('b1000000-0000-0000-0000-000000000003', 'XL', 3);

-- Color: Black, White, Blue, Red
insert into attribute_options (category_attribute_id, value, sort_order) values
  ('b1000000-0000-0000-0000-000000000004', 'Black', 0),
  ('b1000000-0000-0000-0000-000000000004', 'White', 1),
  ('b1000000-0000-0000-0000-000000000004', 'Blue', 2),
  ('b1000000-0000-0000-0000-000000000004', 'Red', 3);

-- Products
insert into products (id, category_id, name, slug, description, base_price, status) values
  ('c1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000003', 'ProBook 15', 'probook-15', 'Professional laptop with stunning display and all-day battery life.', 999.99, 'active'),
  ('c1000000-0000-0000-0000-000000000002', 'a1000000-0000-0000-0000-000000000003', 'UltraSlim Air', 'ultraslim-air', 'Ultra-lightweight laptop for professionals on the go.', 1299.99, 'active'),
  ('c1000000-0000-0000-0000-000000000003', 'a1000000-0000-0000-0000-000000000004', 'Classic Cotton Tee', 'classic-cotton-tee', 'Soft 100% cotton t-shirt. Essential wardrobe staple.', 24.99, 'active'),
  ('c1000000-0000-0000-0000-000000000004', 'a1000000-0000-0000-0000-000000000004', 'Performance Dry-Fit', 'performance-dry-fit', 'Moisture-wicking athletic t-shirt for active lifestyles.', 34.99, 'active');

-- SKUs for ProBook 15
insert into skus (id, product_id, sku_code, price_override, status) values
  ('d1000000-0000-0000-0000-000000000001', 'c1000000-0000-0000-0000-000000000001', 'PROBOOK-8-256', null, 'active'),
  ('d1000000-0000-0000-0000-000000000002', 'c1000000-0000-0000-0000-000000000001', 'PROBOOK-16-512', 1199.99, 'active'),
  ('d1000000-0000-0000-0000-000000000003', 'c1000000-0000-0000-0000-000000000001', 'PROBOOK-32-1TB', 1499.99, 'active');

-- SKU attribute values for ProBook 15
insert into sku_attribute_values (sku_id, category_attribute_id, value) values
  ('d1000000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', '8GB'),
  ('d1000000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000002', '256GB'),
  ('d1000000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000001', '16GB'),
  ('d1000000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000002', '512GB'),
  ('d1000000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000001', '32GB'),
  ('d1000000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000002', '1TB');

-- SKUs for Classic Cotton Tee (4 colors x 4 sizes = 16, showing a subset)
insert into skus (id, product_id, sku_code, status) values
  ('d1000000-0000-0000-0000-000000000010', 'c1000000-0000-0000-0000-000000000003', 'CTEE-BLK-S', 'active'),
  ('d1000000-0000-0000-0000-000000000011', 'c1000000-0000-0000-0000-000000000003', 'CTEE-BLK-M', 'active'),
  ('d1000000-0000-0000-0000-000000000012', 'c1000000-0000-0000-0000-000000000003', 'CTEE-BLU-M', 'active'),
  ('d1000000-0000-0000-0000-000000000013', 'c1000000-0000-0000-0000-000000000003', 'CTEE-WHT-L', 'active');

insert into sku_attribute_values (sku_id, category_attribute_id, value) values
  ('d1000000-0000-0000-0000-000000000010', 'b1000000-0000-0000-0000-000000000003', 'S'),
  ('d1000000-0000-0000-0000-000000000010', 'b1000000-0000-0000-0000-000000000004', 'Black'),
  ('d1000000-0000-0000-0000-000000000011', 'b1000000-0000-0000-0000-000000000003', 'M'),
  ('d1000000-0000-0000-0000-000000000011', 'b1000000-0000-0000-0000-000000000004', 'Black'),
  ('d1000000-0000-0000-0000-000000000012', 'b1000000-0000-0000-0000-000000000003', 'M'),
  ('d1000000-0000-0000-0000-000000000012', 'b1000000-0000-0000-0000-000000000004', 'Blue'),
  ('d1000000-0000-0000-0000-000000000013', 'b1000000-0000-0000-0000-000000000003', 'L'),
  ('d1000000-0000-0000-0000-000000000013', 'b1000000-0000-0000-0000-000000000004', 'White');

-- Custom fields (reporting metadata)
insert into custom_fields (entity_type, entity_id, key, value) values
  ('product', 'c1000000-0000-0000-0000-000000000001', 'supplier', 'TechCorp'),
  ('product', 'c1000000-0000-0000-0000-000000000001', 'tag', 'bestseller'),
  ('product', 'c1000000-0000-0000-0000-000000000003', 'supplier', 'Cotton Co'),
  ('product', 'c1000000-0000-0000-0000-000000000003', 'tag', 'essentials'),
  ('product', 'c1000000-0000-0000-0000-000000000003', 'season', 'SS26');
```

- [ ] **Step 2: Apply seed data**

```bash
supabase db reset
```

This re-runs all migrations and then applies `seed.sql`.

- [ ] **Step 3: Verify via Supabase Studio**

Open `http://127.0.0.1:54323` and check that all tables have data.

- [ ] **Step 4: Commit**

```bash
git add supabase/seed.sql
git commit -m "feat: add seed data for categories, products, and SKUs"
```

---

## Task 13: Nuxt — Layout & API Composable

**Files:**
- Create: `frontend/composables/useApi.ts`
- Create: `frontend/layouts/default.vue`
- Modify: `frontend/app.vue`

- [ ] **Step 1: Create API composable**

Create `frontend/composables/useApi.ts`:

```ts
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function get<T>(path: string): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`)
  }

  async function post<T>(path: string, body: any): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'POST',
      body,
    })
  }

  async function patch<T>(path: string, body: any): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'PATCH',
      body,
    })
  }

  async function del(path: string): Promise<void> {
    await $fetch(`${baseURL}/api/v1${path}`, {
      method: 'DELETE',
    })
  }

  return { get, post, patch, del }
}
```

- [ ] **Step 2: Create default layout**

Create `frontend/layouts/default.vue`:

```vue
<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <NuxtLink to="/" class="text-xl font-bold text-gray-900 tracking-tight">
            FlexShop
          </NuxtLink>
          <nav class="flex items-center gap-6">
            <NuxtLink
              to="/catalog"
              class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
            >
              Catalog
            </NuxtLink>
            <NuxtLink
              to="/cart"
              class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
            >
              Cart
            </NuxtLink>
          </nav>
        </div>
      </div>
    </header>
    <main>
      <slot />
    </main>
    <footer class="bg-white border-t border-gray-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <p class="text-sm text-gray-400 text-center">
          &copy; {{ new Date().getFullYear() }} FlexShop. All rights reserved.
        </p>
      </div>
    </footer>
  </div>
</template>
```

- [ ] **Step 3: Update app.vue to use layout**

Replace `frontend/app.vue` contents with:

```vue
<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
```

- [ ] **Step 4: Verify Nuxt starts**

```bash
cd frontend
npm run dev
```

Expected: Nuxt dev server starts on `http://localhost:3000`. Shows the layout with header/footer.

- [ ] **Step 5: Commit**

```bash
git add frontend/composables/useApi.ts frontend/layouts/default.vue frontend/app.vue
git commit -m "feat: add API composable and default layout"
```

---

## Task 14: Nuxt — Homepage

**Files:**
- Create: `frontend/pages/index.vue`
- Create: `frontend/components/CategoryCard.vue`

- [ ] **Step 1: Create CategoryCard component**

Create `frontend/components/CategoryCard.vue`:

```vue
<template>
  <NuxtLink
    :to="`/catalog/${category.slug}`"
    class="group block bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
  >
    <div class="aspect-[4/3] bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center">
      <span class="text-4xl text-primary-300 group-hover:scale-110 transition-transform">
        {{ icon }}
      </span>
    </div>
    <div class="p-4">
      <h3 class="font-semibold text-gray-900 group-hover:text-primary-600 transition-colors">
        {{ category.name }}
      </h3>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
const props = defineProps<{
  category: { id: string; name: string; slug: string }
  icon?: string
}>()

const icon = props.icon || '+'
</script>
```

- [ ] **Step 2: Create homepage**

Create `frontend/pages/index.vue`:

```vue
<template>
  <div>
    <!-- Hero -->
    <section class="bg-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 text-center">
        <h1 class="text-4xl font-bold text-gray-900 sm:text-5xl">
          Welcome to FlexShop
        </h1>
        <p class="mt-4 text-lg text-gray-600 max-w-2xl mx-auto">
          Quality products, flexible choices. Browse our catalog and find exactly what you need.
        </p>
        <NuxtLink
          to="/catalog"
          class="mt-8 inline-block bg-primary-600 text-white px-8 py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors"
        >
          Browse Catalog
        </NuxtLink>
      </div>
    </section>

    <!-- Categories -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h2 class="text-2xl font-bold text-gray-900 mb-6">Shop by Category</h2>
      <div v-if="categories.length" class="grid grid-cols-2 md:grid-cols-4 gap-6">
        <CategoryCard
          v-for="cat in categories"
          :key="cat.id"
          :category="cat"
        />
      </div>
      <p v-else class="text-gray-500">No categories available.</p>
    </section>
  </div>
</template>

<script setup lang="ts">
const { get } = useApi()

const { data: categories } = await useAsyncData('categories', () =>
  get<Array<{ id: string; name: string; slug: string }>>('/categories?parent_id=null')
)
</script>
```

- [ ] **Step 3: Verify homepage renders**

With both the Go API and Nuxt running:

```bash
# Terminal 1: Go API
cd api && SUPABASE_SERVICE_ROLE_KEY=<your-key> go run cmd/server/main.go

# Terminal 2: Nuxt
cd frontend && npm run dev
```

Open `http://localhost:3000`. Expected: Hero section and category cards (Electronics, Clothing) rendered via SSR.

- [ ] **Step 4: Commit**

```bash
git add frontend/pages/index.vue frontend/components/CategoryCard.vue
git commit -m "feat: add homepage with category grid (SSR)"
```

---

## Task 15: Nuxt — Category Page

**Files:**
- Create: `frontend/pages/catalog/index.vue`
- Create: `frontend/pages/catalog/[slug].vue`
- Create: `frontend/components/ProductCard.vue`

- [ ] **Step 1: Create ProductCard component**

Create `frontend/components/ProductCard.vue`:

```vue
<template>
  <NuxtLink
    :to="`/product/${product.slug}`"
    class="group block bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
  >
    <div class="aspect-square bg-gray-100 flex items-center justify-center">
      <img
        v-if="product.images?.length"
        :src="product.images[0]"
        :alt="product.name"
        class="w-full h-full object-cover"
      />
      <span v-else class="text-gray-300 text-5xl">&#9744;</span>
    </div>
    <div class="p-4">
      <h3 class="font-semibold text-gray-900 group-hover:text-primary-600 transition-colors truncate">
        {{ product.name }}
      </h3>
      <p v-if="product.description" class="mt-1 text-sm text-gray-500 line-clamp-2">
        {{ product.description }}
      </p>
      <p class="mt-2 text-lg font-bold text-gray-900">
        ${{ product.base_price.toFixed(2) }}
      </p>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{
  product: {
    id: string
    name: string
    slug: string
    description?: string
    base_price: number
    images?: string[]
  }
}>()
</script>
```

- [ ] **Step 2: Create catalog index page**

Create `frontend/pages/catalog/index.vue`:

```vue
<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">All Categories</h1>
    <div v-if="categories?.length" class="grid grid-cols-2 md:grid-cols-4 gap-6">
      <CategoryCard
        v-for="cat in categories"
        :key="cat.id"
        :category="cat"
      />
    </div>
    <p v-else class="text-gray-500">No categories found.</p>
  </div>
</template>

<script setup lang="ts">
const { get } = useApi()

const { data: categories } = await useAsyncData('all-categories', () =>
  get<Array<{ id: string; name: string; slug: string }>>('/categories')
)
</script>
```

- [ ] **Step 3: Create category detail page**

Create `frontend/pages/catalog/[slug].vue`:

```vue
<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div v-if="category">
      <!-- Breadcrumb -->
      <nav class="text-sm text-gray-500 mb-6">
        <NuxtLink to="/catalog" class="hover:text-gray-700">Catalog</NuxtLink>
        <span class="mx-2">/</span>
        <span class="text-gray-900">{{ category.name }}</span>
      </nav>

      <h1 class="text-3xl font-bold text-gray-900 mb-2">{{ category.name }}</h1>

      <!-- Subcategories -->
      <div v-if="subcategories?.length" class="mb-8">
        <h2 class="text-lg font-semibold text-gray-700 mb-3">Subcategories</h2>
        <div class="flex flex-wrap gap-3">
          <NuxtLink
            v-for="sub in subcategories"
            :key="sub.id"
            :to="`/catalog/${sub.slug}`"
            class="px-4 py-2 bg-white border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:border-primary-300 hover:text-primary-600 transition-colors"
          >
            {{ sub.name }}
          </NuxtLink>
        </div>
      </div>

      <!-- Products grid -->
      <div v-if="products?.length" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
        <ProductCard
          v-for="product in products"
          :key="product.id"
          :product="product"
        />
      </div>
      <p v-else class="text-gray-500">No products in this category yet.</p>
    </div>
    <div v-else>
      <p class="text-gray-500">Category not found.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { get } = useApi()

const slug = route.params.slug as string

const { data: category } = await useAsyncData(`category-${slug}`, () =>
  get<{ id: string; name: string; slug: string }>(`/categories/${slug}`)
)

const { data: subcategories } = await useAsyncData(`subcats-${slug}`, () =>
  category.value
    ? get<Array<{ id: string; name: string; slug: string }>>(`/categories?parent_id=${category.value.id}`)
    : Promise.resolve([])
)

const { data: products } = await useAsyncData(`products-${slug}`, () =>
  category.value
    ? get<Array<any>>(`/products?category_id=${category.value.id}`)
    : Promise.resolve([])
)
</script>
```

- [ ] **Step 4: Verify category pages render**

Open `http://localhost:3000/catalog` — should show all categories.
Open `http://localhost:3000/catalog/electronics` — should show Electronics with subcategory "Laptops".
Open `http://localhost:3000/catalog/laptops` — should show laptop products.

- [ ] **Step 5: Commit**

```bash
git add frontend/pages/catalog/ frontend/components/ProductCard.vue
git commit -m "feat: add catalog and category pages with product grid (SSR)"
```

---

## Task 16: Nuxt — Product Detail Page

**Files:**
- Create: `frontend/pages/product/[slug].vue`
- Create: `frontend/components/SkuSelector.vue`
- Create: `frontend/components/PriceDisplay.vue`

- [ ] **Step 1: Create PriceDisplay component**

Create `frontend/components/PriceDisplay.vue`:

```vue
<template>
  <span class="text-2xl font-bold text-gray-900">
    ${{ displayPrice.toFixed(2) }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  basePrice: number
  priceOverride?: number | null
}>()

const displayPrice = computed(() => props.priceOverride ?? props.basePrice)
</script>
```

- [ ] **Step 2: Create SkuSelector component**

Create `frontend/components/SkuSelector.vue`:

```vue
<template>
  <div class="space-y-4">
    <div v-for="attr in attributes" :key="attr.id">
      <label class="block text-sm font-medium text-gray-700 mb-1">
        {{ attr.name }}
      </label>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="option in attr.options"
          :key="option"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium border transition-colors',
            selectedValues[attr.name] === option
              ? 'border-primary-500 bg-primary-50 text-primary-700'
              : isOptionAvailable(attr.name, option)
                ? 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'
                : 'border-gray-100 bg-gray-50 text-gray-300 cursor-not-allowed'
          ]"
          :disabled="!isOptionAvailable(attr.name, option)"
          @click="selectOption(attr.name, option)"
        >
          {{ option }}
        </button>
      </div>
    </div>

    <div v-if="selectedSku" class="pt-2">
      <p class="text-sm text-gray-500">SKU: {{ selectedSku.sku_code }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface SKU {
  id: string
  sku_code: string
  price_override: number | null
  attribute_values: Array<{
    category_attribute_id: string
    value: string
  }>
}

interface Attribute {
  id: string
  name: string
  options: string[]
}

const props = defineProps<{
  skus: SKU[]
  attributes: Attribute[]
}>()

const emit = defineEmits<{
  select: [sku: SKU | null]
}>()

const selectedValues = reactive<Record<string, string>>({})

function selectOption(attrName: string, value: string) {
  selectedValues[attrName] = value
  emit('select', selectedSku.value)
}

function isOptionAvailable(attrName: string, option: string): boolean {
  // Check if any SKU matches current selections + this option
  return props.skus.some(sku => {
    const attrMap = buildAttrMap(sku)
    if (attrMap[attrName] !== option) return false
    for (const [name, val] of Object.entries(selectedValues)) {
      if (name !== attrName && attrMap[name] !== val) return false
    }
    return true
  })
}

function buildAttrMap(sku: SKU): Record<string, string> {
  const map: Record<string, string> = {}
  for (const av of sku.attribute_values) {
    const attr = props.attributes.find(a => a.id === av.category_attribute_id)
    if (attr) map[attr.name] = av.value
  }
  return map
}

const selectedSku = computed<SKU | null>(() => {
  if (Object.keys(selectedValues).length !== props.attributes.length) return null
  return props.skus.find(sku => {
    const attrMap = buildAttrMap(sku)
    return Object.entries(selectedValues).every(([name, val]) => attrMap[name] === val)
  }) || null
})
</script>
```

- [ ] **Step 3: Create product detail page**

Create `frontend/pages/product/[slug].vue`:

```vue
<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div v-if="product">
      <!-- Breadcrumb -->
      <nav class="text-sm text-gray-500 mb-6">
        <NuxtLink to="/catalog" class="hover:text-gray-700">Catalog</NuxtLink>
        <span class="mx-2">/</span>
        <span class="text-gray-900">{{ product.name }}</span>
      </nav>

      <div class="grid md:grid-cols-2 gap-12">
        <!-- Image -->
        <div class="aspect-square bg-gray-100 rounded-xl flex items-center justify-center overflow-hidden">
          <img
            v-if="product.images?.length"
            :src="product.images[0]"
            :alt="product.name"
            class="w-full h-full object-cover"
          />
          <span v-else class="text-gray-300 text-8xl">&#9744;</span>
        </div>

        <!-- Details -->
        <div>
          <h1 class="text-3xl font-bold text-gray-900">{{ product.name }}</h1>
          <p v-if="product.description" class="mt-3 text-gray-600 leading-relaxed">
            {{ product.description }}
          </p>

          <!-- Price -->
          <div class="mt-6">
            <PriceDisplay
              :base-price="product.base_price"
              :price-override="selectedSku?.price_override"
            />
          </div>

          <!-- SKU Selector -->
          <div v-if="skus?.length && attributes?.length" class="mt-8">
            <SkuSelector
              :skus="skus"
              :attributes="formattedAttributes"
              @select="onSkuSelect"
            />
          </div>

          <!-- Add to Cart (placeholder for Plan 2) -->
          <button
            class="mt-8 w-full bg-primary-600 text-white py-3 px-6 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
            :disabled="!selectedSku"
          >
            {{ selectedSku ? 'Add to Cart' : 'Select options' }}
          </button>
        </div>
      </div>
    </div>
    <div v-else>
      <p class="text-gray-500">Product not found.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { get } = useApi()

const slug = route.params.slug as string

const { data: product } = await useAsyncData(`product-${slug}`, () =>
  get<any>(`/products/${slug}`)
)

const { data: skus } = await useAsyncData(`skus-${slug}`, async () => {
  if (!product.value) return []
  return get<any[]>(`/products/${product.value.id}/skus`)
})

const { data: attributes } = await useAsyncData(`attrs-${slug}`, async () => {
  if (!product.value) return []
  return get<any[]>(`/categories/${product.value.category_id}/attributes`)
})

const formattedAttributes = computed(() => {
  if (!attributes.value) return []
  return attributes.value.map((attr: any) => ({
    id: attr.id,
    name: attr.name,
    options: attr.options?.map((o: any) => o.value) || [],
  }))
})

const selectedSku = ref<any>(null)

function onSkuSelect(sku: any) {
  selectedSku.value = sku
}
</script>
```

- [ ] **Step 4: Verify product detail page**

Open `http://localhost:3000/product/probook-15`.
Expected: Product name, description, price, SKU selector with RAM and Storage dropdowns. Selecting 16GB + 512GB should show price $1,199.99 and SKU code PROBOOK-16-512.

Open `http://localhost:3000/product/classic-cotton-tee`.
Expected: T-shirt with Size and Color selectors.

- [ ] **Step 5: Commit**

```bash
git add frontend/pages/product/ frontend/components/SkuSelector.vue frontend/components/PriceDisplay.vue
git commit -m "feat: add product detail page with SKU selector (SSR)"
```

---

## Task 17: Final Integration & Cleanup

**Files:**
- Modify: `PROGRESS.md`

- [ ] **Step 1: Run all Go tests**

```bash
cd api
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 2: Run full stack manually**

```bash
# Terminal 1
supabase start && supabase db reset

# Terminal 2
cd api && SUPABASE_SERVICE_ROLE_KEY=<key> go run cmd/server/main.go

# Terminal 3
cd frontend && npm run dev
```

Verify:
- `http://localhost:3000` — Homepage with categories
- `http://localhost:3000/catalog/laptops` — Laptop products
- `http://localhost:3000/product/probook-15` — Product detail with SKU selector
- `http://localhost:8080/health` — API health check

- [ ] **Step 3: Update PROGRESS.md**

Update `PROGRESS.md` to mark Plan 1 milestones as complete:

```markdown
- [x] Plan 1: Foundation & Catalog — implementation plan written
- [x] Plan 1: Foundation & Catalog — implemented
```

And update Plan 1 status to `Complete`.

- [ ] **Step 4: Commit and push**

```bash
git add -A
git commit -m "feat: complete Plan 1 — Foundation & Catalog"
git push origin main
```
