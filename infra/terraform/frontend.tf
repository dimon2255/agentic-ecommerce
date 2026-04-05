resource "azurerm_container_app" "frontend" {
  name                         = "eshop-frontend"
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
    target_port      = 3000
    transport        = "http"

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  template {
    min_replicas = 1
    max_replicas = 5

    container {
      name   = "eshop-frontend"
      image  = "${data.azurerm_container_registry.main.login_server}/eshop-frontend:${var.frontend_image_tag}"
      cpu    = 0.5
      memory = "1Gi"

      env {
        name  = "NUXT_PUBLIC_API_BASE"
        value = "https://${local.api_fqdn}"
      }

      env {
        name  = "NUXT_PUBLIC_SUPABASE_URL"
        value = var.supabase_url
      }

      env {
        name  = "NUXT_PUBLIC_SUPABASE_KEY"
        value = var.frontend_supabase_key
      }

      env {
        name  = "NUXT_PUBLIC_STRIPE_KEY"
        value = var.frontend_stripe_key
      }

      env {
        name  = "HOST"
        value = "0.0.0.0"
      }

      env {
        name  = "PORT"
        value = "3000"
      }

      env {
        name  = "NODE_ENV"
        value = "production"
      }

      liveness_probe {
        transport        = "HTTP"
        path             = "/"
        port             = 3000
        interval_seconds = 30
        initial_delay    = 10
      }

      readiness_probe {
        transport        = "HTTP"
        path             = "/"
        port             = 3000
        interval_seconds = 10
        initial_delay    = 5
      }
    }
  }
}
