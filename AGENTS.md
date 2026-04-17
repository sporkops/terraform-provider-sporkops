# Agent guide for the Spork Terraform provider

You are likely an AI agent helping a user manage Spork uptime monitoring as Terraform code. This file tells you how to write valid HCL and avoid common mistakes.

## When to use this provider vs alternatives

- **Resources that should live in git, be reviewed in PRs, and applied via CI** → use this provider.
- **One-off interactive changes** → use the [Spork CLI](https://github.com/sporkops/cli) with `--json`.
- **Custom Go services** → use the [spork-go SDK](https://github.com/sporkops/spork-go).

## Provider setup

```hcl
terraform {
  required_providers {
    spork = {
      source  = "sporkops/sporkops"
      version = "~> 0.6"
    }
  }
}

provider "spork" {}
```

The provider authenticates from `SPORK_API_KEY`. **Do not write the API key into a config that gets committed.** If you must template it in, use a `sensitive = true` variable backed by an env var or secret store.

## Resources

- `spork_monitor` — uptime monitor (HTTP / DNS / TCP / SSL / keyword / ping)
- `spork_alert_channel` — email / webhook / Slack / Discord / Teams / PagerDuty / Telegram / Google Chat
- `spork_status_page` — public status page
- `spork_member` — organization membership

## Data sources

`spork_monitor`, `spork_monitors`, `spork_alert_channel`, `spork_alert_channels`, `spork_status_page`, `spork_status_pages`, `spork_member`, `spork_members`, `spork_organization`.

## Minimum-viable monitor

```hcl
resource "spork_monitor" "homepage" {
  name     = "Homepage"
  type     = "http"
  target   = "https://example.com"
  interval = 60
}
```

The `target` field's accepted format depends on `type`:

| `type`    | `target` format                                  |
|-----------|--------------------------------------------------|
| `http`    | full URL with scheme                             |
| `keyword` | full URL with scheme                             |
| `ssl`     | full URL with scheme                             |
| `dns`     | bare hostname (e.g. `example.com`)               |
| `ping`    | bare hostname                                    |
| `tcp`     | `host:port` (e.g. `db.example.com:5432`)         |

`interval` is in seconds. Read `/docs/resources/monitor.md` in this repo for the full schema and per-type required fields (e.g. keyword monitors need a `keyword` argument).

## Common patterns

### Monitor + alert channel + attachment

```hcl
resource "spork_alert_channel" "oncall" {
  name = "On-Call Email"
  type = "email"
  config = {
    to = "oncall@example.com"
  }
}

resource "spork_monitor" "homepage" {
  name              = "Homepage"
  type              = "http"
  target            = "https://example.com"
  interval          = 60
  alert_channel_ids = [spork_alert_channel.oncall.id]
}
```

### Many monitors from a list

```hcl
variable "urls" {
  type = list(string)
}

resource "spork_monitor" "site" {
  for_each = toset(var.urls)
  name     = each.key
  type     = "http"
  target   = each.key
  interval = 60
}
```

Use `for_each` over `count` so reordering the input list doesn't churn state.

### Status page driven by monitor IDs

```hcl
resource "spork_status_page" "public" {
  name      = "Acme Status"
  slug      = "acme"
  is_public = true
  components = [
    for m in spork_monitor.site : {
      monitor_id   = m.id
      display_name = m.name
    }
  ]
}
```

## Conventions agents should follow

- **Always set `name`** on every resource — humans recognize them by name in the dashboard.
- **Read `/examples/resources/spork_<type>/`** before generating HCL for an unfamiliar resource. Each subdirectory has runnable, type-specific examples (e.g. `spork_monitor/dns.tf`, `keyword.tf`, `headers.tf`).
- **Run `terraform plan` before `apply`** and show the user the diff. Never apply destructive changes (deletes, replacements) without explicit user confirmation.
- **Don't hardcode API keys.** Read from env or a secret store.
- **Use `sensitive = true`** on any variable that holds a key, webhook URL with a secret, etc.

## Documentation

Full per-resource and per-data-source docs:

- In this repo: `/docs/resources/*.md` and `/docs/data-sources/*.md`
- On the Terraform Registry: <https://registry.terraform.io/providers/sporkops/sporkops/latest/docs>

The registry docs are generated from the in-repo markdown.

## Reporting issues

File bugs at <https://github.com/sporkops/terraform-provider-sporkops/issues>. Include the provider version, the resource block, the full error, and (if relevant) `terraform plan` output.
