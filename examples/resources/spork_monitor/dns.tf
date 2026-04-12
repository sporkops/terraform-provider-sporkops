# DNS monitor — resolves the target hostname periodically.
# Target must be a bare hostname or IP (no scheme, no path, no port).
resource "spork_monitor" "dns_check" {
  name     = "DNS Resolution"
  target   = "example.com"
  type     = "dns"
  interval = 300
}
