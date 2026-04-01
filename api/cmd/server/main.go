package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/checkout"
	"github.com/dimon2255/agentic-ecommerce/api/internal/config"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
	"github.com/dimon2255/agentic-ecommerce/api/internal/requestid"
	stripeClient "github.com/dimon2255/agentic-ecommerce/api/pkg/stripe"
	"github.com/dimon2255/agentic-ecommerce/api/pkg/supabase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey, cfg.Supabase.Timeout)
	auth := middleware.NewAuthMiddleware(cfg.Supabase.JWTSecret)

	categoryHandler := catalog.NewCategoryHandler(db)
	attributeHandler := catalog.NewAttributeHandler(db)
	productHandler := catalog.NewProductHandler(db)
	skuHandler := catalog.NewSKUHandler(db)
	customFieldHandler := catalog.NewCustomFieldHandler(db)
	cartHandler := cart.NewCartHandler(db)
	stripePayments := stripeClient.NewClient(cfg.Stripe.SecretKey, cfg.Stripe.WebhookSecret)
	checkoutHandler := checkout.NewCheckoutHandler(db, stripePayments, cfg.Checkout.PaymentCurrency, cfg.Checkout.WebhookMaxBodySize)

	r := chi.NewRouter()

	r.Use(requestid.Middleware)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
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

		// Cart routes — OptionalAuth so both guests and users can access
		r.Group(func(r chi.Router) {
			r.Use(auth.OptionalAuth)
			r.Mount("/cart", cartHandler.Routes())
		})

		// Checkout and order routes
		r.Route("/checkout", func(r chi.Router) {
			r.Use(auth.OptionalAuth)
			r.Mount("/", checkoutHandler.Routes())
		})
		r.Route("/orders", func(r chi.Router) {
			r.Use(auth.OptionalAuth)
			r.Mount("/", checkoutHandler.OrderRoutes())
		})
	})

	// Stripe webhook — outside /api/v1, no auth (uses signature verification)
	r.Mount("/stripe/webhook", checkoutHandler.WebhookRoutes())

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("API server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
