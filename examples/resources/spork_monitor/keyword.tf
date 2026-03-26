resource "spork_monitor" "keyword_check" {
  target       = "https://example.com/health"
  name         = "Health Check Keyword"
  type         = "keyword"
  keyword      = "healthy"
  keyword_type = "exists"
  interval     = 300
}
