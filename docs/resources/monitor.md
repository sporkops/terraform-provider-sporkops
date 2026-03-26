---
page_title: "spork_monitor Resource"
description: |-
  Manages a Spork uptime monitor.
---

# spork_monitor

Manages a Spork uptime monitor that periodically checks a target URL and reports its status.

## Example Usage

```hcl
resource "spork_monitor" "website" {
  target   = "https://example.com"
  name     = "Production Website"
  type     = "http"
  method   = "GET"
  interval = 60
}
```

### Monitor with All Options

```hcl
resource "spork_monitor" "api" {
  target            = "https://api.example.com/health"
  name              = "API Health Check"
  type              = "http"
  method            = "POST"
  expected_status   = 201
  interval          = 300
  timeout           = 15
  regions           = ["us-central1", "europe-west1"]
  alert_channel_ids = [spork_alert_channel.slack.id]
  tags              = ["production", "api"]
  paused            = false
  headers = {
    "Authorization" = "Bearer token"
    "Accept"        = "application/json"
  }
  body = "{\"check\": true}"
}
```

## Argument Reference

- `target` (Required, String) — The URL to monitor. Must start with `http://` or `https://`.
- `name` (Required, String) — A human-readable name for the monitor (1-100 characters).
- `type` (Optional, String) — Monitor type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`. Default: `http`.
- `method` (Optional, String) — HTTP method to use for checks. One of: `GET`, `HEAD`, `POST`, `PUT`. Default: `GET`.
- `expected_status` (Optional, Number) — Expected HTTP status code (100-599). Default: `200`.
- `interval` (Optional, Number) — Check interval in seconds (60-3600). Default: `60`.
- `timeout` (Optional, Number) — Timeout in seconds for each check (5-120). Default: `30`.
- `regions` (Optional, List of String) — Regions to check from. Available: `us-central1`, `europe-west1`. Default: `["us-central1"]`.
- `alert_channel_ids` (Optional, List of String) — IDs of alert channels to notify on status changes.
- `tags` (Optional, List of String) — Tags for organizing monitors.
- `paused` (Optional, Boolean) — Whether the monitor is paused. Default: `false`.
- `headers` (Optional, Map of String) — Custom HTTP request headers to send with each check. Applicable to HTTP-based monitor types.
- `body` (Optional, String) — HTTP request body to send with each check. Applicable to HTTP-based monitor types.
- `keyword` (Optional, String) — The keyword to search for in the response body. **Required** when `type = "keyword"`.
- `keyword_type` (Optional, String) — Whether the keyword must exist or must not exist in the response. One of: `exists`, `not_exists`. Default: `exists`. Only used when `type = "keyword"`.
- `ssl_warn_days` (Optional, Number) — Number of days before SSL certificate expiry to trigger a warning. Only used when `type = "ssl"`. Default: `30`.

## Type-Specific Requirements

- **keyword monitors** (`type = "keyword"`): `keyword` is required. `keyword_type` defaults to `"exists"` and accepts `"exists"` or `"not_exists"`.
- **ssl monitors** (`type = "ssl"`): `ssl_warn_days` is optional and defaults to `30`.
- **HTTP monitors** (`type = "http"`): `headers` and `body` are optional and can be used with any HTTP monitor type.

## Attribute Reference

- `id` (String) — The unique identifier of the monitor.
- `status` (String) — Current monitor status: `up`, `down`, `degraded`, `paused`, or `pending`.

## Import

Monitors can be imported using their ID:

```shell
terraform import spork_monitor.website mon_abc123
```
