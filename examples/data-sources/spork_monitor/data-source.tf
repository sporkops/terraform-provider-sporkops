# Look up a monitor by name
data "spork_monitor" "api" {
  name = "API Health"
}

output "api_monitor_status" {
  value = data.spork_monitor.api.status
}
