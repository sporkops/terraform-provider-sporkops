---
page_title: "spork_monitor Data Source"
description: |-
  Fetches a Spork uptime monitor by ID.
---

# spork_monitor (Data Source)

Fetches a [Spork](https://sporkops.com) uptime monitor by ID.

## Example Usage

```hcl
data "spork_monitor" "website" {
  id = "mon_abc123"
}

output "monitor_status" {
  value = data.spork_monitor.website.status
}
```

### Read a Keyword Monitor

```hcl
data "spork_monitor" "keyword_check" {
  id = "mon_def456"
}

output "keyword" {
  value = data.spork_monitor.keyword_check.keyword
}

output "keyword_type" {
  value = data.spork_monitor.keyword_check.keyword_type
}
```

### Read an SSL Monitor

```hcl
data "spork_monitor" "ssl_check" {
  id = "mon_ghi789"
}

output "ssl_warn_days" {
  value = data.spork_monitor.ssl_check.ssl_warn_days
}
```

## Argument Reference

- `id` (Required, String) — The unique identifier of the monitor.

## Attribute Reference

- `target` (String) — The URL being monitored.
- `name` (String) — A human-readable name for the monitor.
- `type` (String) — Monitor type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`.
- `method` (String) — HTTP method used for checks. One of: `GET`, `HEAD`, `POST`, `PUT`.
- `interval` (Number) — Check interval in seconds.
- `timeout` (Number) — Timeout in seconds for each check.
- `expected_status` (Number) — Expected HTTP status code.
- `paused` (Boolean) — Whether the monitor is paused.
- `status` (String) — Current monitor status: `up`, `down`, `degraded`, `paused`, or `pending`.
- `regions` (List of String) — Regions the monitor checks from.
- `alert_channel_ids` (List of String) — IDs of alert channels notified on status changes.
- `tags` (List of String) — Tags for organizing monitors.
- `headers` (Map of String) — Custom HTTP request headers sent with each check.
- `body` (String) — HTTP request body sent with each check.
- `keyword` (String) — The keyword searched for in the response body. Set when `type = "keyword"`.
- `keyword_type` (String) — Whether the keyword must exist or not. One of: `exists`, `not_exists`.
- `ssl_warn_days` (Number) — Days before SSL certificate expiry to trigger a warning. Set when `type = "ssl"`.
- `created_at` (String) — Timestamp when the monitor was created.
- `updated_at` (String) — Timestamp when the monitor was last updated.
