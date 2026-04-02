package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Supabase  SupabaseConfig  `mapstructure:"supabase"`
	Stripe    StripeConfig    `mapstructure:"stripe"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Checkout  CheckoutConfig  `mapstructure:"checkout"`
	Assistant AssistantConfig `mapstructure:"assistant"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

type SupabaseConfig struct {
	URL            string        `mapstructure:"url"`
	ServiceRoleKey string        `mapstructure:"service_role_key"`
	JWTSecret      string        `mapstructure:"jwt_secret"`
	JWTIssuer      string        `mapstructure:"jwt_issuer"`
	JWTAudience    string        `mapstructure:"jwt_audience"`
	Timeout        time.Duration `mapstructure:"timeout"`
}

type StripeConfig struct {
	SecretKey     string `mapstructure:"secret_key"`
	WebhookSecret string `mapstructure:"webhook_secret"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	MaxAge         int      `mapstructure:"max_age"`
}

type CheckoutConfig struct {
	PaymentCurrency    string `mapstructure:"payment_currency"`
	WebhookMaxBodySize int64  `mapstructure:"webhook_max_body_size"`
}

type AssistantConfig struct {
	AnthropicAPIKey string `mapstructure:"anthropic_api_key"`
	VoyageAPIKey    string `mapstructure:"voyage_api_key"`
	Model           string `mapstructure:"model"`
	EmbeddingModel  string `mapstructure:"embedding_model"`
}

func Load(configPaths ...string) (*Config, error) {
	// Load .env file if present (check current dir and parent dir)
	for _, path := range []string{".env", "../.env"} {
		if f, err := os.Open(path); err == nil {
			gotenv.Apply(f)
			f.Close()
			break
		}
	}

	v := viper.New()

	// Defaults
	v.SetDefault("server.port", 9090)
	v.SetDefault("server.request_timeout", "30s")
	v.SetDefault("supabase.url", "http://127.0.0.1:54321")
	v.SetDefault("supabase.timeout", "10s")
	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:3001"})
	v.SetDefault("cors.max_age", 300)
	v.SetDefault("checkout.payment_currency", "usd")
	v.SetDefault("checkout.webhook_max_body_size", 65536)
	v.SetDefault("assistant.model", "claude-sonnet-4-5-20241022")
	v.SetDefault("assistant.embedding_model", "voyage-3-large")

	// YAML config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if len(configPaths) > 0 {
		for _, p := range configPaths {
			v.AddConfigPath(p)
		}
	} else {
		v.AddConfigPath(".")
	}

	// Environment variables — must explicitly bind nested keys for Unmarshal
	v.SetEnvPrefix("ESHOP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind all config keys to env vars so Unmarshal picks them up
	for _, key := range []string{
		"server.port",
		"server.request_timeout",
		"supabase.url",
		"supabase.service_role_key",
		"supabase.jwt_secret",
		"supabase.jwt_issuer",
		"supabase.jwt_audience",
		"supabase.timeout",
		"stripe.secret_key",
		"stripe.webhook_secret",
		"cors.allowed_origins",
		"cors.max_age",
		"checkout.payment_currency",
		"checkout.webhook_max_body_size",
		"assistant.anthropic_api_key",
		"assistant.voyage_api_key",
		"assistant.model",
		"assistant.embedding_model",
	} {
		v.BindEnv(key)
	}

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Validate required fields
	if cfg.Supabase.ServiceRoleKey == "" {
		return nil, fmt.Errorf("supabase.service_role_key is required (set ESHOP_SUPABASE_SERVICE_ROLE_KEY)")
	}
	if cfg.Supabase.JWTSecret == "" {
		return nil, fmt.Errorf("supabase.jwt_secret is required (set ESHOP_SUPABASE_JWT_SECRET)")
	}

	return &cfg, nil
}
