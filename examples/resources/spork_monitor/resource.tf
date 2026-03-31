# Basic HTTP monitor
resource "spork_monitor" "api" {
  name   = "API Health"
  target = "https://api.example.com/health"
}

# Monitor with custom settings
resource "spork_monitor" "website" {
  name            = "Website"
  target          = "https://example.com"
  interval        = 300
  timeout         = 60
  expected_status = 200
  regions         = ["us-central1", "europe-west1"]

  alert_channel_ids = [spork_alert_channel.email.id]
  tags              = ["production", "critical"]
}

# Keyword monitoring
resource "spork_monitor" "api_keyword" {
  name         = "API Response Check"
  target       = "https://api.example.com/status"
  type         = "keyword"
  keyword      = "healthy"
  keyword_type = "exists"
}
