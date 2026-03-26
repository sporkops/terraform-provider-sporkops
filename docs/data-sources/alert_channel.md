---
page_title: "spork_alert_channel Data Source"
description: |-
  Fetches a Spork alert channel by ID.
---

# spork_alert_channel (Data Source)

Fetches a [Spork](https://sporkops.com) alert channel by ID.

## Example Usage

```hcl
data "spork_alert_channel" "slack" {
  id = "ch_abc123"
}

output "alert_channel_name" {
  value = data.spork_alert_channel.slack.name
}
```

## Argument Reference

- `id` (Required, String) — The unique identifier of the alert channel.

## Attribute Reference

- `name` (String) — A friendly name for the alert channel.
- `type` (String) — The channel type: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, or `googlechat`.
- `config` (Map of String) — Channel-specific configuration as key-value pairs.
- `verified` (Boolean) — Whether the channel has been verified. Relevant for `email` type.
- `created_at` (String) — Timestamp when the alert channel was created.
