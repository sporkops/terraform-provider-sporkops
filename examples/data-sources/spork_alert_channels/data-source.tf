data "spork_alert_channels" "all" {}

output "channel_count" {
  value = length(data.spork_alert_channels.all.alert_channels)
}
