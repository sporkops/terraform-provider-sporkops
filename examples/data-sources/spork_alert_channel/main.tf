data "spork_alert_channel" "slack" {
  id = "ch_abc123"
}

output "alert_channel_name" {
  value = data.spork_alert_channel.slack.name
}
