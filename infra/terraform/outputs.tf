output "api_fqdn" {
  description = "FQDN of the API container app"
  value       = azurerm_container_app.api.ingress[0].fqdn
}

output "frontend_fqdn" {
  description = "FQDN of the frontend container app"
  value       = azurerm_container_app.frontend.ingress[0].fqdn
}

output "frontend_url" {
  description = "Public URL of the frontend"
  value       = "https://${azurerm_container_app.frontend.ingress[0].fqdn}"
}
