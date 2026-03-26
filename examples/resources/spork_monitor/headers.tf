resource "spork_monitor" "with_headers" {
  target = "https://api.example.com/health"
  name   = "API Health Check"
  headers = {
    "Authorization" = "Bearer test-token"
    "Accept"        = "application/json"
  }
}
