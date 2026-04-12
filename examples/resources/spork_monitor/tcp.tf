# TCP monitor — opens a TCP connection to the target host:port.
# Target must be in host:port format (no scheme, no path).
resource "spork_monitor" "tcp_check" {
  name     = "TCP Connectivity"
  target   = "example.com:443"
  type     = "tcp"
  interval = 120
}
