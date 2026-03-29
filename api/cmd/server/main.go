package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/dimon2255/agentic-ecommerce/api/internal/cart"
	"github.com/dimon2255/agentic-ecommerce/api/internal/catalog"
	"github.com/dimon2255/agentic-ecommerce/api/internal/middleware"
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
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("SUPABASE_JWT_SECRET is required")
	}

	db := supabase.NewClient(supabaseURL, supabaseKey)
	auth := middleware.NewAuthMiddleware(jwtSecret)

	categoryHandler := catalog.NewCategoryHandler(db)
	attributeHandler := catalog.NewAttributeHandler(db)
	productHandler := catalog.NewProductHandler(db)
	skuHandler := catalog.NewSKUHandler(db)
	customFieldHandler := catalog.NewCustomFieldHandler(db)
	cartHandler := cart.NewCartHandler(db)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Session-ID"},
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

		// Cart routes — OptionalAuth so both guests and users can access
		r.Group(func(r chi.Router) {
			r.Use(auth.OptionalAuth)
			r.Mount("/cart", cartHandler.Routes())
		})
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("API server listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
