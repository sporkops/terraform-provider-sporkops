---
page_title: "spork_alert_channel Resource"
description: |-
  Manages a Spork alert channel for uptime notifications.
---

# spork_alert_channel

Manages a Spork alert channel that receives notifications when a monitor detects downtime.

## Example Usage

### Email Channel

```hcl
resource "spork_alert_channel" "oncall" {
  name = "On-Call Team"
  type = "email"
  config = {
    to = "oncall@example.com"
  }
}
```

### Slack Webhook

```hcl
resource "spork_alert_channel" "slack" {
  name = "Slack Alerts"
  type = "slack"
  config = {
    url = "https://hooks.slack.com/services/T00/B00/xxx"
  }
}
```

### Generic Webhook

```hcl
resource "spork_alert_channel" "webhook" {
  name = "Custom Webhook"
  type = "webhook"
  config = {
    url = "https://api.example.com/alerts"
  }
}
```

### PagerDuty

```hcl
resource "spork_alert_channel" "pagerduty" {
  name = "PagerDuty Oncall"
  type = "pagerduty"
  config = {
    integration_key = "your-pagerduty-integration-key"
  }
}
```

## Argument Reference

- `name` (Required, String) — A friendly name for the alert channel.
- `type` (Required, String) — The channel type: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, or `googlechat`. Changing this forces a new resource.
- `config` (Required, Map of String) — Channel-specific configuration as key-value pairs:
  - **email**: `{to = "addr"}`
  - **webhook/slack/discord/teams/googlechat**: `{url = "webhook_url"}`
  - **pagerduty**: `{integration_key = "key"}`
  - **telegram**: `{bot_token = "token", chat_id = "id"}`

## Attribute Reference

- `id` (String) — The unique identifier of the alert channel.

## Important Notes

**Webhook signing secret**: For `webhook` channel types, `config.secret` is only returned at creation time. After creation, it is redacted from API responses and will not appear in state refreshes. Store the secret securely at creation time.

**Email verification**: Email channels require verification. After creating an email channel, a verification email is sent to the configured address. The channel will not deliver alerts until the email address has been verified.

## Import

Alert channels can be imported using their ID:

```shell
terraform import spork_alert_channel.oncall ch_abc123
```
