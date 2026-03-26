resource "spork_monitor" "ssl_check" {
  target        = "https://example.com"
  name          = "SSL Certificate"
  type          = "ssl"
  ssl_warn_days = 14
}
