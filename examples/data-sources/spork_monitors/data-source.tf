# List all monitors
data "spork_monitors" "all" {}

output "monitor_count" {
  value = length(data.spork_monitors.all.monitors)
}
