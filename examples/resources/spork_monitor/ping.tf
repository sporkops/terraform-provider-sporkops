# Ping monitor — sends ICMP echo requests to the target.
# Target must be a bare hostname or IP (no scheme, no path, no port).
resource "spork_monitor" "ping_check" {
  name     = "ICMP Ping"
  target   = "example.com"
  type     = "ping"
  interval = 120
}
