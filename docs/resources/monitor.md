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

### DNS, TCP, and Ping Monitors

Network-level monitors use a different target format — a bare hostname for DNS
and ping, `host:port` for TCP. URL schemes (`http://`, `https://`) and paths
are rejected at `terraform plan` time.

```hcl
resource "spork_monitor" "dns_check" {
  name   = "DNS Resolution"
  target = "example.com"
  type   = "dns"
}

resource "spork_monitor" "tcp_check" {
  name   = "TCP Connectivity"
  target = "example.com:443"
  type   = "tcp"
}

resource "spork_monitor" "ping_check" {
  name   = "ICMP Ping"
  target = "example.com"
  type   = "ping"
}
```

## Argument Reference

- `target` (Required, String) — The target to monitor. Format depends on `type`:
  - `http`, `keyword`, `ssl`: URL starting with `http://` or `https://` (e.g., `https://example.com`).
  - `dns`, `ping`: bare hostname or IP (e.g., `example.com`). No URL scheme, path, or port.
  - `tcp`: `host:port` (e.g., `example.com:443`).
- `name` (Required, String) — A human-readable name for the monitor (1-100 characters).
- `type` (Optional, String) — Monitor type. One of: `http`, `ssl`, `dns`, `keyword`, `tcp`, `ping`. Default: `http`.
- `method` (Optional, String) — HTTP method to use for checks. One of: `GET`, `HEAD`, `POST`, `PUT`. Default: `GET`.
- `expected_status` (Optional, Number) — Expected HTTP status code (100-599). Default: `200`.
- `interval` (Optional, Number) — Check interval in seconds (60-86400, must be a multiple of 60). Default: `60`.
- `timeout` (Optional, Number) — Timeout in seconds for each check (5-120). Default: `14`.
- `regions` (Optional, List of String) — Regions to check from. Available: `us-central1`, `europe-west1`. Default: `["us-central1"]`.
- `alert_channel_ids` (Optional, List of String) — IDs of alert channels to notify on status changes.
- `tags` (Optional, List of String) — Tags for organizing monitors.
- `paused` (Optional, Boolean) — Whether the monitor is paused. Default: `false`.
- `headers` (Optional, Map of String) — Custom HTTP request headers to send with each check. Applicable to HTTP-based monitor types.
- `body` (Optional, String) — HTTP request body to send with each check. Applicable to HTTP-based monitor types.
- `keyword` (Optional, String) — The keyword to search for in the response body. **Required** when `type = "keyword"`.
- `keyword_type` (Optional, String) — Whether the keyword must exist or must not exist in the response. One of: `exists`, `not_exists`. Default: `exists`. Only used when `type = "keyword"`.
- `ssl_warn_days` (Optional, Number) — Number of days before SSL certificate expiry to trigger a warning. Only used when `type = "ssl"`. Default: `14`.

## Type-Specific Requirements

- **HTTP monitors** (`type = "http"`): `target` must start with `http://` or `https://`. `headers` and `body` are optional.
- **keyword monitors** (`type = "keyword"`): `target` must start with `http://` or `https://`. `keyword` is required. `keyword_type` defaults to `"exists"` and accepts `"exists"` or `"not_exists"`.
- **SSL monitors** (`type = "ssl"`): `target` must start with `http://` or `https://`. `ssl_warn_days` is optional and defaults to `14`.
- **DNS monitors** (`type = "dns"`): `target` must be a bare hostname or IP (e.g. `example.com`). URL schemes, paths, and embedded ports are rejected.
- **TCP monitors** (`type = "tcp"`): `target` must be `host:port` (e.g. `example.com:443`). URL schemes and paths are rejected; the port must be an integer from 1 to 65535.
- **Ping monitors** (`type = "ping"`): `target` must be a bare hostname or IP (e.g. `example.com`). URL schemes, paths, and embedded ports are rejected.

## Attribute Reference

- `id` (String) — The unique identifier of the monitor.
- `status` (String) — Current monitor status: `up`, `down`, `degraded`, `paused`, or `pending`.

## Import

Monitors can be imported using their ID:

```shell
terraform import spork_monitor.website mon_abc123
```
