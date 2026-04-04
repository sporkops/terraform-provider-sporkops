---
page_title: "spork_alert_channels Data Source"
description: |-
  Fetches all Spork alert channels with optional filtering.
---

# spork_alert_channels (Data Source)

Fetches all [Spork](https://sporkops.com) alert channels with optional filtering by type.

## Example Usage

```hcl
data "spork_alert_channels" "all" {}

output "channel_count" {
  value = length(data.spork_alert_channels.all.alert_channels)
}
```

### Filter by type

```hcl
data "spork_alert_channels" "slack_only" {
  type = "slack"
}
```

## Schema

### Optional

- `type` (String) — Filter alert channels by type. One of: `email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, `googlechat`.

### Read-Only

- `alert_channels` — List of alert channels matching the filter. Each entry contains:
  - `id` — Alert channel ID.
  - `name` — Alert channel name.
  - `type` — Channel type (`email`, `webhook`, `slack`, `discord`, `teams`, `pagerduty`, `telegram`, `googlechat`).
  - `config` — Channel-specific configuration as key-value pairs.
  - `verified` — Whether the channel has been verified (relevant for email type).
  - `created_at` — Creation timestamp.
