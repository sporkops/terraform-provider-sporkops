variable "monitor_auth_token" {
  type        = string
  sensitive   = true
  description = "Bearer token sent with each check. Supply via TF_VAR_monitor_auth_token or a secrets backend; never commit."
}

resource "spork_monitor" "with_headers" {
  target = "https://api.example.com/health"
  name   = "API Health Check"
  headers = {
    "Authorization" = "Bearer ${var.monitor_auth_token}"
    "Accept"        = "application/json"
  }
}
