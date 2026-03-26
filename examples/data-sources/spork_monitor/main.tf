data "spork_monitor" "website" {
  id = "mon_abc123"
}

output "monitor_status" {
  value = data.spork_monitor.website.status
}
