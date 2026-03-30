# Email alert channel
resource "spork_alert_channel" "email" {
  name = "Team Email"
  type = "email"
  config = {
    to = "oncall@example.com"
  }
}

# Webhook alert channel
resource "spork_alert_channel" "webhook" {
  name = "PagerDuty Webhook"
  type = "webhook"
  config = {
    url = "https://events.pagerduty.com/integration/abc123/enqueue"
  }
}

# Slack alert channel
resource "spork_alert_channel" "slack" {
  name = "Slack Alerts"
  type = "slack"
  config = {
    url = "https://hooks.slack.com/services/T00/B00/xxx"
  }
}
