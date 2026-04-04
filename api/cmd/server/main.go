package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/dimon2255/agentic-ecommerce/api/internal/assistant"
	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/checkout"
	"github.com/dimon2255/agentic-ecommerce/api/internal/config"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/internal/requestid"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/circuitbreaker"
	stripeClient "github.com/dimon2255/agentic-ecommerce/api/pkg/stripe"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/telemetry"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// Bootstrap JSON logger (replaced with trace-aware handler after telemetry init)
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(jsonHandler))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize OpenTelemetry (no-op when telemetry.otlp_endpoint is empty)
	ctx := context.Background()
	telemetryShutdown, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:  cfg.Telemetry.ServiceName,
		OTLPEndpoint: cfg.Telemetry.OTLPEndpoint,
	})
	if err != nil {
		slog.Error("failed to init telemetry", "error", err)
		os.Exit(1)
	}
	defer telemetryShutdown(ctx)

	// Upgrade to trace-aware slog handler (injects trace_id/span_id into log records)
	slog.SetDefault(slog.New(telemetry.NewTracedHandler(jsonHandler)))

	db := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.Timeout)
	auth := middleware.NewAuthMiddleware(cfg.Supabase.JWTSecret, cfg.Supabase.JWTIssuer, cfg.Supabase.JWTAudience, cfg.Supabase.URL)

	// Rate limiters
	apiLimiter := middleware.NewRateLimiter(100, time.Minute)
	checkoutLimiter := middleware.NewRateLimiter(20, time.Minute)
	webhookLimiter := middleware.NewRateLimiter(50, time.Minute)
	webhookReplay := middleware.NewWebhookReplayGuard(10000)
	assistantLimiter := middleware.NewAssistantRateLimiter(
		cfg.Assistant.RateLimit.UserMessagesPerHour,
		cfg.Assistant.RateLimit.UserBurstPerMinute,
		cfg.Assistant.RateLimit.GuestMessagesPerHour,
		cfg.Assistant.RateLimit.GuestBurstPerMinute,
	)

	catalogRepo := catalog.NewSupabaseRepository(db)
	catalogSvc := catalog.NewService(catalogRepo)
	categoryHandler := catalog.NewCategoryHandler(catalogSvc)
	attributeHandler := catalog.NewAttributeHandler(catalogSvc)
	productHandler := catalog.NewProductHandler(catalogSvc)
	skuHandler := catalog.NewSKUHandler(catalogSvc)
	customFieldHandler := catalog.NewCustomFieldHandler(catalogSvc)
	cartRepo := cart.NewSupabaseRepository(db)
	cartSvc := cart.NewService(cartRepo)
	cartHandler := cart.NewCartHandler(cartSvc)
	stripePayments := stripeClient.NewClient(cfg.Stripe.SecretKey, cfg.Stripe.WebhookSecret)
	checkoutRepo := checkout.NewSupabaseRepository(db)
	checkoutSvc := checkout.NewService(checkoutRepo, stripePayments, cfg.Checkout.PaymentCurrency)
	checkoutHandler := checkout.NewCheckoutHandler(checkoutSvc, stripePayments, cfg.Checkout.WebhookMaxBodySize)

	// AI Shopping Assistant
	voyageClient := voyage.NewClient(cfg.Assistant.VoyageAPIKey, cfg.Assistant.EmbeddingModel)
	anthropicClient := anthropic.NewClient(cfg.Assistant.AnthropicAPIKey, cfg.Assistant.Model)
	cb := circuitbreaker.New(
		cfg.Assistant.Cost.CircuitBreakerThreshold,
		cfg.Assistant.Cost.CircuitBreakerWindow,
		cfg.Assistant.Cost.CircuitBreakerOpenDur,
	)
	anthropicClient.SetCircuitBreaker(cb)
	assistantRepo := assistant.NewSupabaseRepository(db)
	assistantSvc := assistant.NewService(assistantRepo, voyageClient, anthropicClient, catalogSvc, cartSvc, assistant.ServiceConfig{
		Model:            cfg.Assistant.Model,
		DailyBudgetCents: cfg.Assistant.Cost.DailyBudgetCents,
	})
	assistantHandler := assistant.NewHandler(assistantSvc)

	r := chi.NewRouter()

	r.Use(requestid.Middleware)
	r.Use(otelhttp.NewMiddleware("eshop-api"))
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	// Note: Timeout middleware is applied per-group (not globally) so SSE streaming routes
	// can bypass http.TimeoutHandler which doesn't support http.Flusher.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Session-ID"},
		AllowCredentials: true,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	_ = webhookReplay // wired into checkout handler for webhook dedup

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(apiLimiter.Middleware)

		// Standard routes with request timeout
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(cfg.Server.RequestTimeout))

			r.Mount("/categories", categoryHandler.Routes())
			r.Route("/categories/{categoryId}/attributes", func(r chi.Router) {
				r.Mount("/", attributeHandler.Routes())
			})
			r.Mount("/products", productHandler.Routes())
			r.Route("/products/{productId}/skus", func(r chi.Router) {
				r.Mount("/", skuHandler.Routes())
			})
			r.Mount("/custom-fields", customFieldHandler.Routes())

			// Cart routes — OptionalAuth so both guests and users can access
			r.Group(func(r chi.Router) {
				r.Use(auth.OptionalAuth)
				r.Mount("/cart", cartHandler.Routes())
			})

			// AI Assistant routes (non-streaming) — supports guests with limited tools
			r.Route("/assistant", func(r chi.Router) {
				r.Use(auth.OptionalAuth)
				r.Use(assistantLimiter.Middleware)
				r.Mount("/", assistantHandler.Routes())
			})

			// Checkout and order routes
			r.Route("/checkout", func(r chi.Router) {
				r.Use(checkoutLimiter.Middleware)
				r.Use(auth.OptionalAuth)
				r.Mount("/", checkoutHandler.Routes())
			})
			r.Route("/orders", func(r chi.Router) {
				r.Use(auth.OptionalAuth)
				r.Mount("/", checkoutHandler.OrderRoutes())
			})
		})

		// SSE streaming route — NO timeout middleware (handler manages its own 2-min context deadline)
		r.Route("/assistant/stream", func(r chi.Router) {
			r.Use(auth.OptionalAuth)
			r.Use(assistantLimiter.Middleware)
			r.Post("/", assistantHandler.StreamRoute())
		})
	})

	// Stripe webhook — outside /api/v1, no auth (uses signature verification)
	r.Group(func(r chi.Router) {
		r.Use(webhookLimiter.Middleware)
		r.Use(middleware.Timeout(cfg.Server.RequestTimeout))
		r.Mount("/stripe/webhook", checkoutHandler.WebhookRoutes())
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	// Graceful shutdown on SIGTERM/SIGINT
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		slog.Info("received signal, shutting down", "signal", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info("API server listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped gracefully")
}
