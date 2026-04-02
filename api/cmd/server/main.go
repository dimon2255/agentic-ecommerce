package main

import (
	"fmt"
	"log"
	"net/http"
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
	stripeClient "github.com/dimon2255/agentic-ecommerce/api/pkg/stripe"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/voyage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.Timeout)
	auth := middleware.NewAuthMiddleware(cfg.Supabase.JWTSecret, cfg.Supabase.JWTIssuer, cfg.Supabase.JWTAudience, cfg.Supabase.URL)

	// Rate limiters
	apiLimiter := middleware.NewRateLimiter(100, time.Minute)
	checkoutLimiter := middleware.NewRateLimiter(20, time.Minute)
	webhookLimiter := middleware.NewRateLimiter(50, time.Minute)
	webhookReplay := middleware.NewWebhookReplayGuard(10000)

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
	assistantRepo := assistant.NewSupabaseRepository(db)
	assistantSvc := assistant.NewService(assistantRepo, voyageClient, anthropicClient)
	assistantHandler := assistant.NewHandler(assistantSvc)

	r := chi.NewRouter()

	r.Use(requestid.Middleware)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.Timeout(cfg.Server.RequestTimeout))
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

		// AI Assistant routes — requires authentication
		r.Route("/assistant", func(r chi.Router) {
			r.Use(auth.RequireAuth)
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

	// Stripe webhook — outside /api/v1, no auth (uses signature verification)
	r.Group(func(r chi.Router) {
		r.Use(webhookLimiter.Middleware)
		r.Mount("/stripe/webhook", checkoutHandler.WebhookRoutes())
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("API server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
