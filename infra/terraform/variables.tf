# --- Environment ---

variable "environment" {
  description = "Deployment environment (staging or production)"
  type        = string

  validation {
    condition     = contains(["staging", "production"], var.environment)
    error_message = "Environment must be 'staging' or 'production'."
  }
}

# --- Existing infrastructure references ---

variable "resource_group_name" {
  description = "Name of the existing resource group"
  type        = string
}

variable "aca_environment_name" {
  description = "Name of the existing ACA environment"
  type        = string
}

variable "acr_name" {
  description = "Name of the existing Azure Container Registry"
  type        = string
}

variable "managed_identity_name" {
  description = "Name of the existing user-assigned managed identity"
  type        = string
}

# --- Image tags ---

variable "api_image_tag" {
  description = "Docker image tag for the API container"
  type        = string
  default     = "latest"
}

variable "frontend_image_tag" {
  description = "Docker image tag for the frontend container"
  type        = string
  default     = "latest"
}

# --- Secrets (sensitive) ---

variable "eshop_supabase_service_role_key" {
  description = "Supabase service role key (bypasses RLS)"
  type        = string
  sensitive   = true
}

variable "eshop_supabase_jwt_secret" {
  description = "Supabase JWT verification secret"
  type        = string
  sensitive   = true
}

variable "eshop_stripe_secret_key" {
  description = "Stripe secret API key"
  type        = string
  sensitive   = true
}

variable "eshop_stripe_webhook_secret" {
  description = "Stripe webhook signing secret"
  type        = string
  sensitive   = true
}

variable "eshop_assistant_anthropic_api_key" {
  description = "Anthropic API key for AI assistant"
  type        = string
  sensitive   = true
  default     = ""
}

variable "eshop_assistant_voyage_api_key" {
  description = "Voyage API key for embeddings"
  type        = string
  sensitive   = true
  default     = ""
}

# --- Runtime config (non-secret) ---

variable "supabase_url" {
  description = "Supabase Cloud project URL"
  type        = string
}

variable "supabase_jwt_issuer" {
  description = "Expected JWT issuer claim (optional)"
  type        = string
  default     = ""
}

variable "supabase_jwt_audience" {
  description = "Expected JWT audience claim (optional)"
  type        = string
  default     = ""
}

variable "frontend_supabase_key" {
  description = "Supabase anon/publishable key (safe for client-side)"
  type        = string
}

variable "frontend_stripe_key" {
  description = "Stripe publishable key (safe for client-side)"
  type        = string
  default     = ""
}
