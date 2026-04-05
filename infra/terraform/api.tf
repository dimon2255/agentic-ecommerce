resource "azurerm_container_app" "api" {
  name                         = local.api_app_name
  container_app_environment_id = data.azurerm_container_app_environment.main.id
  resource_group_name          = data.azurerm_resource_group.main.name
  revision_mode                = "Single"

  identity {
    type         = "UserAssigned"
    identity_ids = [data.azurerm_user_assigned_identity.main.id]
  }

  registry {
    server   = data.azurerm_container_registry.main.login_server
    identity = data.azurerm_user_assigned_identity.main.id
  }

  ingress {
    external_enabled = true
    target_port      = 9090
    transport        = "http"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  secret {
    name  = "supabase-service-role-key"
    value = var.eshop_supabase_service_role_key
  }

  secret {
    name  = "supabase-jwt-secret"
    value = var.eshop_supabase_jwt_secret
  }

  secret {
    name  = "stripe-secret-key"
    value = var.eshop_stripe_secret_key
  }

  secret {
    name  = "stripe-webhook-secret"
    value = var.eshop_stripe_webhook_secret
  }

  secret {
    name  = "anthropic-api-key"
    value = var.eshop_assistant_anthropic_api_key
  }

  secret {
    name  = "voyage-api-key"
    value = var.eshop_assistant_voyage_api_key
  }

  template {
    min_replicas = 1
    max_replicas = 3

    container {
      name   = "eshop-api"
      image  = "${data.azurerm_container_registry.main.login_server}/eshop-api:${var.api_image_tag}"
      cpu    = 0.25
      memory = "0.5Gi"

      # Non-secret environment variables
      env {
        name  = "ESHOP_SERVER_PORT"
        value = "9090"
      }

      env {
        name  = "ESHOP_SUPABASE_URL"
        value = var.supabase_url
      }

      env {
        name  = "ESHOP_SUPABASE_JWT_ISSUER"
        value = var.supabase_jwt_issuer
      }

      env {
        name  = "ESHOP_SUPABASE_JWT_AUDIENCE"
        value = var.supabase_jwt_audience
      }

      env {
        name  = "ESHOP_CORS_ALLOWED_ORIGINS"
        value = "https://${local.frontend_fqdn}"
      }

      # Secret references
      env {
        name        = "ESHOP_SUPABASE_SERVICE_ROLE_KEY"
        secret_name = "supabase-service-role-key"
      }

      env {
        name        = "ESHOP_SUPABASE_JWT_SECRET"
        secret_name = "supabase-jwt-secret"
      }

      env {
        name        = "ESHOP_STRIPE_SECRET_KEY"
        secret_name = "stripe-secret-key"
      }

      env {
        name        = "ESHOP_STRIPE_WEBHOOK_SECRET"
        secret_name = "stripe-webhook-secret"
      }

      env {
        name        = "ESHOP_ASSISTANT_ANTHROPIC_API_KEY"
        secret_name = "anthropic-api-key"
      }

      env {
        name        = "ESHOP_ASSISTANT_VOYAGE_API_KEY"
        secret_name = "voyage-api-key"
      }

      startup_probe {
        transport               = "HTTP"
        path                    = "/health"
        port                    = 9090
        interval_seconds        = 3
        failure_count_threshold = 10
      }

      liveness_probe {
        transport               = "HTTP"
        path                    = "/health"
        port                    = 9090
        interval_seconds        = 30
        failure_count_threshold = 3
      }

      readiness_probe {
        transport               = "HTTP"
        path                    = "/health"
        port                    = 9090
        interval_seconds        = 10
        failure_count_threshold = 3
      }
    }
  }
}
