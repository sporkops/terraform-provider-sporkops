resource "spork_status_page" "main" {
  name  = "Acme Status"
  slug  = "acme"
  theme = "light"

  components {
    monitor_id   = spork_monitor.api.id
    display_name = "API"
    order        = 0
  }

  components {
    monitor_id   = spork_monitor.web.id
    display_name = "Website"
    order        = 1
  }
}
