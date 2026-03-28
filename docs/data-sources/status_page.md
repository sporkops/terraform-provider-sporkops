---
page_title: "spork_status_page Data Source"
description: |-
  Fetches a Spork status page by ID or name.
---

# spork_status_page (Data Source)

Fetches a [Spork](https://sporkops.com) status page by ID or name.

## Example Usage

```hcl
data "spork_status_page" "main" {
  id = "sp_abc123"
}

output "status_page_slug" {
  value = data.spork_status_page.main.slug
}
```

### Look Up by Name

```hcl
data "spork_status_page" "main" {
  name = "Acme Status"
}
```

## Argument Reference

- `id` (Optional, String) — The unique identifier of the status page. Specify either `id` or `name`.
- `name` (Optional, String) — The name of the status page. Specify either `id` or `name`.

## Attribute Reference

- `slug` — URL-safe slug for the status page.
- `components` — List of components displayed on the status page.
  - `id` — Component ID.
  - `monitor_id` — The monitor ID.
  - `display_name` — Display name on the status page.
  - `description` — Component description.
  - `order` — Display order.
- `custom_domain` — Custom domain, if configured.
- `domain_status` — Custom domain verification status.
- `theme` — Color theme (`light` or `dark`).
- `accent_color` — Accent color hex code.
- `logo_url` — Logo URL.
- `is_public` — Whether the status page is publicly accessible.
- `created_at` — Creation timestamp.
- `updated_at` — Last update timestamp.
