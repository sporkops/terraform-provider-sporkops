---
page_title: "spork_monitors Data Source"
description: |-
  Fetches all Spork uptime monitors with optional filtering.
---

# spork_monitors (Data Source)

Fetches all [Spork](https://sporkops.com) uptime monitors with optional filtering by type or status.

## Example Usage

```hcl
# List all monitors
data "spork_monitors" "all" {}

output "monitor_count" {
  value = length(data.spork_monitors.all.monitors)
}
```

### Filter by type

```hcl
data "spork_monitors" "http_only" {
  type = "http"
}
```

### Filter by status

```hcl
data "spork_monitors" "down" {
  status = "down"
}
```

## Schema

### Optional

- `type` (String) — Filter monitors by type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`.
- `status` (String) — Filter monitors by status. One of: `up`, `down`, `degraded`, `paused`, `pending`.

### Read-Only

- `monitors` — List of monitors matching the filters. Each entry contains:
  - `id` — Monitor ID.
  - `target` — The URL being monitored.
  - `name` — Monitor name.
  - `type` — Monitor type (`http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`).
  - `method` — HTTP method used for checks (`GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`).
  - `interval` — Check interval in seconds.
  - `timeout` — Timeout in seconds for each check.
  - `expected_status` — Expected HTTP status code.
  - `paused` — Whether the monitor is paused.
  - `status` — Current monitor status (`up`, `down`, `degraded`, `paused`, `pending`).
  - `regions` — Regions the monitor checks from (`us-central1`, `europe-west1`).
  - `alert_channel_ids` — IDs of alert channels notified on status changes.
  - `tags` — Tags for organizing monitors.
  - `headers` — Custom HTTP request headers sent with each check.
  - `body` — HTTP request body sent with each check.
  - `keyword` — The keyword searched for in the response body (when type is `keyword`).
  - `keyword_type` — Whether the keyword must exist or not (`exists`, `not_exists`).
  - `ssl_warn_days` — Days before SSL certificate expiry to trigger a warning (when type is `ssl`).
  - `created_at` — Creation timestamp.
  - `updated_at` — Last update timestamp.
