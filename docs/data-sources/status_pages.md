---
page_title: "spork_status_pages Data Source"
description: |-
  Fetches all Spork status pages.
---

# spork_status_pages (Data Source)

Fetches all [Spork](https://sporkops.com) status pages.

## Example Usage

```hcl
data "spork_status_pages" "all" {}

output "page_count" {
  value = length(data.spork_status_pages.all.status_pages)
}
```

## Attribute Reference

- `status_pages` — List of status pages. Each entry contains:
  - `id` — Status page ID.
  - `name` — Status page name.
  - `slug` — URL-safe slug.
  - `components` — List of components (see `spork_status_page` data source for details).
  - `custom_domain` — Custom domain, if configured.
  - `domain_status` — Custom domain verification status.
  - `theme` — Color theme (`light` or `dark`).
  - `accent_color` — Accent color hex code.
  - `logo_url` — Logo URL.
  - `is_public` — Whether the status page is publicly accessible.
  - `created_at` — Creation timestamp.
  - `updated_at` — Last update timestamp.
