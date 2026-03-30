---
page_title: "spork_status_page Resource"
description: |-
  Manages a Spork status page.
---

# spork_status_page

Manages a [Spork](https://sporkops.com) public status page with components, custom domains, and branding.

## Example Usage

```hcl
resource "spork_status_page" "main" {
  name = "Acme Status"
  slug = "acme"
}
```

### Status Page with Components

```hcl
resource "spork_monitor" "api" {
  target = "https://api.example.com/health"
  name   = "API"
  type   = "http"
}

resource "spork_monitor" "web" {
  target = "https://example.com"
  name   = "Website"
  type   = "http"
}

resource "spork_status_page" "main" {
  name  = "Acme Status"
  slug  = "acme"
  theme = "dark"

  components {
    monitor_id   = spork_monitor.api.id
    display_name = "API"
    order        = 0
  }

  components {
    monitor_id   = spork_monitor.web.id
    display_name = "Website"
    order        = 1
  }
}
```

### Status Page with Custom Domain

```hcl
resource "spork_status_page" "main" {
  name          = "Acme Status"
  slug          = "acme"
  custom_domain = "status.example.com"
  accent_color  = "#0066ff"
  logo_url      = "https://cdn.example.com/logo.png"
}
```

~> **Note:** Custom domains require a CNAME record pointing to `status.sporkops.com`. After setting a custom domain, `domain_status` will be `pending` until DNS propagation completes.

## Argument Reference

- `name` (Required, String) — The name of the status page (1-100 characters).
- `slug` (Required, String) — URL-safe slug for the status page (2-63 lowercase alphanumeric characters or hyphens). Used in the URL: `https://<slug>.status.sporkops.com`.
- `components` (Optional, List of Object) — Components displayed on the status page. Each component maps a monitor to a display name.
  - `monitor_id` (Required, String) — The ID of the monitor to display.
  - `display_name` (Required, String) — The name shown on the status page.
  - `description` (Optional, String) — A description of the component.
  - `order` (Optional, Number) — Display order (0-based).
- `custom_domain` (Optional, String) — Custom domain for the status page (e.g. `status.example.com`). Requires a CNAME record pointing to `status.sporkops.com`.
- `theme` (Optional, String) — Color theme. One of: `light`, `dark`, `blue`, `midnight`. Default: `light`.
- `accent_color` (Optional, String) — Accent color as a hex code (e.g. `#ff0000`).
- `font_family` (Optional, String) — Font family for the status page. One of: `system`, `sans-serif`, `serif`, `monospace`. Default: `system`.
- `header_style` (Optional, String) — Header style for the status page. One of: `default`, `banner`, `minimal`. Default: `default`.
- `logo_url` (Optional, String) — URL of the logo. Must be an `https://` URL.
- `is_public` (Optional, Boolean) — Whether the status page is publicly accessible. Default: `true`.

## Attribute Reference

- `id` — The unique identifier of the status page.
- `domain_status` — Verification status of the custom domain: `pending`, `active`, or `failed`.
- `components[].id` — The unique identifier of each component.
- `created_at` — Timestamp when the status page was created.
- `updated_at` — Timestamp when the status page was last updated.

## Import

Status pages can be imported by their ID:

```shell
terraform import spork_status_page.main sp_abc123
```
