terraform {
  required_version = ">= 1.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }

  # State key is injected at init time via -backend-config="key=eshop-<env>.tfstate"
  backend "azurerm" {
    resource_group_name  = "tfstate-rg"
    storage_account_name = "tfstateeshop"
    container_name       = "tfstate"
    use_oidc             = true
  }
}

provider "azurerm" {
  features {}
  use_oidc = true
}

# --- Data sources for existing infrastructure ---

data "azurerm_resource_group" "main" {
  name = var.resource_group_name
}

data "azurerm_container_app_environment" "main" {
  name                = var.aca_environment_name
  resource_group_name = data.azurerm_resource_group.main.name
}

data "azurerm_container_registry" "main" {
  name                = var.acr_name
  resource_group_name = data.azurerm_resource_group.main.name
}

data "azurerm_user_assigned_identity" "main" {
  name                = var.managed_identity_name
  resource_group_name = data.azurerm_resource_group.main.name
}

# Predict FQDNs from ACA environment to avoid circular dependency between apps
locals {
  api_app_name      = "eshop-api-${var.environment}"
  frontend_app_name = "eshop-frontend-${var.environment}"
  api_fqdn          = "${local.api_app_name}.${data.azurerm_container_app_environment.main.default_domain}"
  frontend_fqdn     = "${local.frontend_app_name}.${data.azurerm_container_app_environment.main.default_domain}"
}
