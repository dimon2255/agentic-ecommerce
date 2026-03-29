# E-Shop Platform Design

## Overview

A single-tenant e-commerce platform for selling physical goods of any type. The system features a flexible, category-driven product catalog with SKU variant support, a server-side shopping cart, and a Stripe-powered checkout flow with 3D Secure (3DS) for customer-initiated transactions (CIT).

**Merchant model:** Single-tenant — one team, one shop, no marketplace functionality.

## Technology Stack

| Layer | Technology | Role |
|-------|-----------|------|
| Frontend | Nuxt 3 (Vue, Vite) | SSR for public pages, SPA for authenticated flows |
| Backend | Go | REST API, business logic, Stripe integration |
| Database | PostgreSQL via Supabase | Data storage, auth, migrations, storage |
| Payments | Stripe | PaymentIntent API, Payment Element, 3DS, webhooks |
| Auth | Supabase Auth | Registration, login, JWT tokens |

## Architecture

Four-layer architecture with clear separation:

1. **Browser** — SSR-hydrated catalog pages + SPA cart/checkout + Stripe.js for payment tokenization + Supabase Auth client SDK.
2. **SSR Server (Node.js)** — Nuxt 3 handles server-side rendering for SEO pages (catalog, product detail, category browsing). Static asset serving.
3. **Go API** — REST endpoints organized by domain (catalog, cart, checkout). Orchestrates business logic, validation, and Stripe integration. All data operations go through Supabase REST API (PostgREST), not direct PostgreSQL.
4. **Supabase (PostgreSQL)** — Data storage with Row Level Security, auth, file storage for product images, migrations via Supabase CLI.

### Data Flow

- **SSR pages:** Browser → Nuxt SSR Server → Go API → Supabase
- **SPA pages:** Browser → Go API → Supabase
- **Payments:** Browser → Stripe.js (tokenize) → Go API → Stripe API (charge + 3DS)
- **Webhooks:** Stripe → Go API → Supabase (order status updates)

### Rendering Strategy

- **SSR:** Homepage, category pages, product detail pages — for SEO and fast initial load.
- **SPA:** Cart, checkout, account pages — for snappy interactivity in authenticated flows.
- Nuxt's hybrid rendering configured per-route via `routeRules`.

### Supabase Integration

Go API communicates with Supabase exclusively through its REST API layer:

- **User-scoped queries:** Go passes user JWTs to Supabase, RLS policies enforce access control at the database level.
- **Admin/backend operations:** Go uses the `service_role` key to bypass RLS when needed (e.g., webhook handlers, admin reports).
- **Complex queries:** PostgreSQL functions called via Supabase RPC endpoint (e.g., admin reports with joins and aggregations).
- **Auth:** Supabase Auth handles registration, login, and JWT issuance. Nuxt uses `@nuxtjs/supabase` module. Go validates JWTs from Supabase.
- **Storage:** Supabase Storage for product images. Admin uploads images directly from the frontend to Supabase Storage (using Supabase client SDK with auth). Go API stores the returned public URLs on the product record. Portable to S3 later if needed.
- **Migrations:** `supabase migration` and `supabase db diff` for schema changes, version controlled in `supabase/migrations/`.

**Portability:** If migrating away from Supabase, swap the data access layer in Go and handle auth independently. The architecture does not leak Supabase-specific logic into business logic.

## Data Model

### Categories

```
categories
├── id (uuid, PK)
├── name (text, not null)
├── slug (text, unique, not null)
├── parent_id (uuid, FK → categories.id, nullable — for subcategories)
├── created_at, updated_at
```

Hierarchical categories via self-referencing `parent_id`. Supports trees like "Electronics > Laptops > Gaming Laptops".

### Category Attributes

```
category_attributes
├── id (uuid, PK)
├── category_id (uuid, FK → categories.id)
├── name (text, not null — e.g. "Size", "Color", "Material")
├── type (enum: text, number, enum — determines input type and validation)
├── required (boolean, default false)
├── sort_order (integer)
```

Each category defines its own attributes. These drive business logic — SKU generation, filtering, product forms.

### Attribute Options

```
attribute_options
├── id (uuid, PK)
├── category_attribute_id (uuid, FK → category_attributes.id)
├── value (text, not null — e.g. "S", "M", "L", "XL")
├── sort_order (integer)
```

Enum values for attributes of type `enum`. Only relevant for enum-type attributes.

### Products

```
products
├── id (uuid, PK)
├── category_id (uuid, FK → categories.id)
├── name (text, not null)
├── slug (text, unique, not null)
├── description (text)
├── base_price (numeric, not null — default price before variant overrides)
├── status (enum: draft, active, archived)
├── images (text[] — array of Supabase Storage URLs)
├── created_at, updated_at
```

### SKUs

```
skus
├── id (uuid, PK)
├── product_id (uuid, FK → products.id)
├── sku_code (text, unique, not null — e.g. "TSHIRT-BLU-M")
├── price_override (numeric, nullable — if null, use product.base_price)
├── status (enum: active, inactive)
├── created_at, updated_at
```

Each unique variant combination is a SKU. SKUs are the purchasable unit — cart items reference SKUs, not products.

### SKU Attribute Values

```
sku_attribute_values
├── id (uuid, PK)
├── sku_id (uuid, FK → skus.id)
├── category_attribute_id (uuid, FK → category_attributes.id)
├── value (text, not null — e.g. "Blue", "M")
```

Links a SKU to its specific attribute values. A SKU for a blue medium t-shirt has two rows: color=Blue, size=M.

### Custom Fields

```
custom_fields
├── id (uuid, PK)
├── entity_type (enum: product, sku)
├── entity_id (uuid, not null)
├── key (text, not null — e.g. "supplier", "season", "tag")
├── value (text, not null)
```

Key-value metadata for reporting and display only. No business logic depends on custom fields. Queryable for admin reports and filtering.

### Carts

```
carts
├── id (uuid, PK)
├── user_id (uuid, FK → auth.users.id, nullable — null for guest carts)
├── session_id (text, not null — ties guest carts to browser session)
├── status (enum: active, merged, expired)
├── created_at, updated_at
```

One active cart per user or session. On login, guest cart merges into user cart.

### Cart Items

```
cart_items
├── id (uuid, PK)
├── cart_id (uuid, FK → carts.id)
├── sku_id (uuid, FK → skus.id)
├── quantity (integer, not null, > 0)
├── unit_price (numeric, not null — snapshot at add-time)
├── created_at, updated_at
```

Price captured at add-time. Re-validated against current SKU price at checkout start — user notified if price changed.

### Orders

```
orders
├── id (uuid, PK)
├── user_id (uuid, FK → auth.users.id, nullable)
├── status (enum: draft, pending, paid, shipped, completed, cancelled)
├── email (text, not null)
├── shipping_address (jsonb, not null)
├── subtotal (numeric, not null)
├── total (numeric, not null)
├── stripe_payment_intent_id (text, unique — idempotency key, prevents double charges)
├── created_at, updated_at
```

Order lifecycle: draft (checkout started) → pending (payment submitted) → paid (webhook confirmed) → shipped → completed. Cancellation possible before shipping.

### Order Items

```
order_items
├── id (uuid, PK)
├── order_id (uuid, FK → orders.id)
├── sku_id (uuid, FK → skus.id)
├── product_name (text — snapshot)
├── sku_code (text — snapshot)
├── quantity (integer, not null)
├── unit_price (numeric, not null — snapshot at checkout)
```

All fields snapshotted at checkout time. Order history remains accurate even if products change or are deleted later.

### Page Views

```
page_views
├── id (uuid, PK)
├── product_id (uuid, FK → products.id)
├── session_id (text)
├── viewed_at (timestamptz, default now())
```

Tracks product page hits for the admin top-SKUs report.

## Shopping Cart

- **Server-side state:** Go API manages cart. Frontend calls `POST /cart/items`, `PATCH /cart/items/:id`, `DELETE /cart/items/:id`.
- **Guest support:** Anonymous users get a cart tied to `session_id`. On login, guest cart merges into user cart (items combined, duplicates summed).
- **Price snapshots:** `unit_price` captured at add-time. At checkout start, Go API re-validates all prices against current SKU prices and notifies the user of any changes.
- **No inventory hold:** Adding to cart does not reserve stock (inventory tracking is out of scope for v1).
- **Cart expiry:** Stale carts (30+ days inactive) cleaned up by a scheduled background job.

## Checkout & Payment Flow

Seven-step checkout using Stripe PaymentIntent API with 3DS for customer-initiated transactions:

1. **Cart review** — user reviews items. Frontend calls `GET /cart` for current state.
2. **Shipping & contact info** — user enters name, email, shipping address. `POST /checkout/start` creates a draft order, re-validates cart prices.
3. **Create PaymentIntent** — Go API calls `stripe.PaymentIntents.Create()` with amount, currency, `payment_method_types: ["card"]`. Stores `payment_intent_id` on the order. Returns `client_secret` to frontend.
4. **Payment form** — Nuxt renders Stripe Payment Element using `client_secret`. Card data handled entirely by Stripe — never touches the Go API.
5. **Confirm payment + 3DS** — frontend calls `stripe.confirmPayment()`. If 3DS is required, Stripe displays the authentication modal automatically. This is the CIT — customer is present and authenticating.
6. **Webhook confirmation** — Stripe sends `payment_intent.succeeded` to Go API's webhook endpoint. Go verifies the webhook signature, updates order status from `pending` → `paid`. This is the source of truth — frontend confirmation is not trusted.
7. **Order confirmation** — frontend polls or listens for order status update. Displays confirmation page with order details.

### Security

- **PCI compliance:** Card data never touches the Go API. Stripe Elements handles all sensitive data.
- **Webhook as truth:** Order status transitions only via verified Stripe webhooks, not frontend calls.
- **Price re-validation:** At `POST /checkout/start`, Go API compares current SKU prices against cart snapshots. User notified of changes before payment.
- **Idempotency:** `stripe_payment_intent_id` stored on the order. Retries are safe — no double charges.

## Admin Report: Top SKUs (RPC)

A PostgreSQL function called via Supabase RPC for the admin dashboard:

**Function:** `get_top_skus_report(page, page_size, sort_by, sort_dir)`

**Returns:**

| Field | Source |
|-------|--------|
| sku_code | skus |
| product_name | products |
| category_name | categories |
| category_slug | categories |
| tags | custom_fields (key = 'tag') |
| page_views | aggregated from page_views |
| total_sold | aggregated from order_items |
| revenue | sum(quantity * unit_price) from order_items |
| total_count | for pagination metadata |

Joins SKUs, products, categories, custom fields, page views, and order items. Handles sorting, filtering, and pagination server-side. Go API calls via `supabase.rpc('get_top_skus_report', params)` and passes results to the admin frontend.

## Project Structure

```
e-shop/
├── frontend/                    # Nuxt 3 app
│   ├── pages/                   # File-based routing
│   │   ├── index.vue            # Homepage (SSR)
│   │   ├── catalog/
│   │   │   ├── [slug].vue       # Category page (SSR)
│   │   │   └── index.vue        # All categories (SSR)
│   │   ├── product/
│   │   │   └── [slug].vue       # Product detail (SSR)
│   │   ├── cart.vue             # Cart page (SPA)
│   │   ├── checkout.vue         # Checkout flow (SPA)
│   │   └── account/             # User account (SPA)
│   ├── components/              # Reusable Vue components
│   ├── composables/             # Shared logic (useCart, useAuth)
│   ├── layouts/                 # Page layouts (default, checkout)
│   └── nuxt.config.ts
│
├── api/                         # Go backend
│   ├── cmd/server/main.go       # Entry point
│   ├── internal/
│   │   ├── catalog/             # Category, product, SKU handlers + Supabase queries
│   │   ├── cart/                # Cart handlers + Supabase queries
│   │   ├── checkout/            # Order, Stripe integration
│   │   ├── admin/               # Admin report handlers
│   │   └── middleware/          # Auth (JWT validation), CORS, logging
│   ├── pkg/
│   │   ├── supabase/            # Supabase client wrapper (REST + RPC)
│   │   └── stripe/              # Stripe client wrapper
│   └── go.mod
│
├── supabase/
│   ├── migrations/              # SQL migration files
│   ├── functions/               # PostgreSQL functions (e.g. get_top_skus_report)
│   ├── seed.sql                 # Dev seed data
│   └── config.toml              # Supabase project config
│
└── docker-compose.yml           # Local dev (Supabase, Go API, Nuxt)
```

## Scope Boundaries

**In scope (v1):**
- Flexible catalog with category-driven attributes and custom fields
- SKU variant model
- Server-side shopping cart with guest support
- Stripe checkout with 3DS (CIT)
- Supabase Auth (registration, login)
- Admin report (top SKUs via RPC)
- Product image storage via Supabase Storage

**Out of scope (v1):**
- Inventory tracking / stock management
- Shipping rate calculation
- Tax calculation
- Order fulfillment / shipping integration
- Email notifications
- Multi-tenant / marketplace features
- Search (full-text or Algolia) — basic category/attribute filtering only
