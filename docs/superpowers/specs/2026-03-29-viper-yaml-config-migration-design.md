# Viper + YAML Config Migration Design

## Problem

The Go API uses scattered `os.Getenv()` calls in `main.go` with no centralized config management. Several values (CORS origins, Supabase timeout, payment currency, webhook body size) are hardcoded. There is no config file — everything must be passed as env vars.

## Solution

Migrate to [Viper](https://github.com/spf13/viper) with a typed `Config` struct in a dedicated `api/internal/config/` package. Non-secret defaults live in `api/config.yaml`. All env vars use `ESHOP_` prefix for clear ownership.

## Config Struct

```go
// api/internal/config/config.go

type Config struct {
    Server   ServerConfig
    Supabase SupabaseConfig
    Stripe   StripeConfig
    CORS     CORSConfig
    Checkout CheckoutConfig
}

type ServerConfig struct {
    Port int `mapstructure:"port"` // default: 9090
}

type SupabaseConfig struct {
    URL            string        `mapstructure:"url"`              // default: http://127.0.0.1:54321
    ServiceRoleKey string        `mapstructure:"service_role_key"` // required, env only
    JWTSecret      string        `mapstructure:"jwt_secret"`       // required, env only
    Timeout        time.Duration `mapstructure:"timeout"`          // default: 10s
}

type StripeConfig struct {
    SecretKey     string `mapstructure:"secret_key"`     // optional, env only
    WebhookSecret string `mapstructure:"webhook_secret"` // optional, env only
}

type CORSConfig struct {
    AllowedOrigins []string `mapstructure:"allowed_origins"` // default: [http://localhost:3000, http://localhost:3001]
    MaxAge         int      `mapstructure:"max_age"`         // default: 300
}

type CheckoutConfig struct {
    PaymentCurrency    string `mapstructure:"payment_currency"`      // default: usd
    WebhookMaxBodySize int64  `mapstructure:"webhook_max_body_size"` // default: 65536
}
```

## YAML File (`api/config.yaml`)

```yaml
server:
  port: 9090

supabase:
  url: http://127.0.0.1:54321
  timeout: 10s

cors:
  allowed_origins:
    - http://localhost:3000
    - http://localhost:3001
  max_age: 300

checkout:
  payment_currency: usd
  webhook_max_body_size: 65536
```

Secrets are NOT stored in YAML — they come from env vars only.

## Env Var Mapping

All env vars use `ESHOP_` prefix. Viper's `SetEnvPrefix("ESHOP")` + `AutomaticEnv()` maps `ESHOP_SUPABASE_URL` to key `supabase.url`.

| YAML Key | Env Var | Required | Default |
|----------|---------|----------|---------|
| `server.port` | `ESHOP_SERVER_PORT` | No | 9090 |
| `supabase.url` | `ESHOP_SUPABASE_URL` | No | http://127.0.0.1:54321 |
| `supabase.service_role_key` | `ESHOP_SUPABASE_SERVICE_ROLE_KEY` | Yes | - |
| `supabase.jwt_secret` | `ESHOP_SUPABASE_JWT_SECRET` | Yes | - |
| `supabase.timeout` | `ESHOP_SUPABASE_TIMEOUT` | No | 10s |
| `stripe.secret_key` | `ESHOP_STRIPE_SECRET_KEY` | No | - |
| `stripe.webhook_secret` | `ESHOP_STRIPE_WEBHOOK_SECRET` | No | - |
| `cors.allowed_origins` | `ESHOP_CORS_ALLOWED_ORIGINS` | No | localhost:3000,3001 |
| `cors.max_age` | `ESHOP_CORS_MAX_AGE` | No | 300 |
| `checkout.payment_currency` | `ESHOP_CHECKOUT_PAYMENT_CURRENCY` | No | usd |
| `checkout.webhook_max_body_size` | `ESHOP_CHECKOUT_WEBHOOK_MAX_BODY_SIZE` | No | 65536 |

## Loading Strategy (`config.Load()`)

1. Set Viper defaults for all non-secret values
2. Set config file name/path (`config.yaml` in `api/`)
3. `SetEnvPrefix("ESHOP")` + `SetEnvKeyReplacer` (`.` -> `_`)
4. `AutomaticEnv()` to bind all env vars
5. `ReadInConfig()` — non-fatal if file missing (env-only mode works)
6. `Unmarshal` into `Config` struct
7. Validate required fields (`ServiceRoleKey`, `JWTSecret`), return error if missing

## Integration Changes

### `main.go`
- Replace all `os.Getenv()` calls with single `config.Load()` call
- Pass config sub-structs to constructors

### `supabase.NewClient()`
- Change signature to accept `SupabaseConfig` (or at minimum accept timeout param)
- Use `cfg.Timeout` instead of hardcoded `10 * time.Second`

### `checkout.NewHandler()`
- Accept `CheckoutConfig` for payment currency and webhook max body size
- Replace hardcoded `"usd"` and `65536`

### CORS setup in `main.go`
- Use `cfg.CORS.AllowedOrigins` and `cfg.CORS.MaxAge` instead of hardcoded slices

### `.env.example`
- Update to reflect new `ESHOP_`-prefixed var names
- Add all missing vars (JWT secret, Stripe keys, new configurable fields)
- Separate API vars from frontend vars with comments

## Files Modified

- `api/internal/config/config.go` (NEW) — Config struct + Load()
- `api/cmd/server/main.go` — Replace os.Getenv with config.Load()
- `api/pkg/supabase/client.go` — Accept timeout from config
- `api/internal/checkout/handler.go` — Accept currency + max body size from config
- `api/config.yaml` (NEW) — Default config file
- `api/go.mod` — Add viper dependency
- `.env.example` — Update with ESHOP_ prefix and all vars
- `.gitignore` — Ensure `config.local.yaml` or similar is ignored

## Testing

- Existing tests should continue to pass (handlers receive same values)
- Add unit test for `config.Load()` — test defaults, YAML override, env override, required field validation
- Manual test: run API with only env vars (no YAML), with YAML only, with both (env overrides YAML)
