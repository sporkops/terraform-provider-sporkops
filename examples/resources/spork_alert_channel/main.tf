resource "spork_alert_channel" "oncall" {
  name = "On-Call Team"
  type = "email"
  config = {
    to = "oncall@example.com"
  }
}

resource "spork_alert_channel" "slack" {
  name = "Slack Alerts"
  type = "slack"
  config = {
    url = "https://hooks.slack.com/services/T00/B00/xxx"
  }
}
