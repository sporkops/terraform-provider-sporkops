# Basic status page
resource "spork_status_page" "main" {
  name      = "Acme Status"
  slug      = "acme-status"
  is_public = true
  theme     = "light"

  components {
    monitor_id   = spork_monitor.api.id
    display_name = "API"
    order        = 0
  }

  components {
    monitor_id   = spork_monitor.website.id
    display_name = "Website"
    order        = 1
  }
}

# Status page with component groups
resource "spork_status_page" "grouped" {
  name      = "Acme Status"
  slug      = "acme-grouped"
  is_public = true
  theme     = "dark"

  component_groups {
    name  = "Core Services"
    order = 0
  }

  component_groups {
    name  = "Supporting Services"
    order = 1
  }

  components {
    monitor_id   = spork_monitor.api.id
    display_name = "API"
    group        = "Core Services"
    order        = 0
  }

  components {
    monitor_id   = spork_monitor.website.id
    display_name = "Website"
    group        = "Core Services"
    order        = 1
  }

  components {
    monitor_id   = spork_monitor.cdn.id
    display_name = "CDN"
    group        = "Supporting Services"
    order        = 0
  }
}
