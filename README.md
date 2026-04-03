# Terraform Provider for Spork

[![Tests](https://github.com/sporkops/terraform-provider-sporkops/actions/workflows/test.yml/badge.svg)](https://github.com/sporkops/terraform-provider-sporkops/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/sporkops/terraform-provider-sporkops)](https://github.com/sporkops/terraform-provider-sporkops/releases/latest)
[![Terraform Registry](https://img.shields.io/badge/Terraform-Registry-purple.svg)](https://registry.terraform.io/providers/sporkops/sporkops/latest)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL--2.0-blue.svg)](LICENSE)

**Know when your site goes down before your customers do.**

Add uptime monitoring to your Terraform workflow. One resource block, real alerts.

## Quickstart

```hcl
resource "spork_monitor" "website" {
  target = "https://yoursite.com"
  name   = "Production Website"
}
```

```sh
export SPORK_API_KEY="your-api-key"   # get one free at sporkops.com
terraform init && terraform apply
```

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- A Spork account ([sign up free](https://sporkops.com))

## Usage

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

resource "spork_monitor" "website" {
  target   = "https://example.com"
  name     = "Production Website"
  type     = "http"
  method   = "GET"
  interval = 60
}

resource "spork_alert_channel" "oncall" {
  name = "On-Call Email"
  type = "email"
  config = {
    to = "oncall@example.com"
  }
}
```

## Authentication

Set your API key via environment variable:

```shell
export SPORK_API_KEY="your-api-key-here"
```

Or configure it in the provider block:

```hcl
provider "spork" {
  api_key = "your-api-key-here"
}
```

Generate an API key from the Spork dashboard at [sporkops.com/settings/api-keys](https://sporkops.com/settings/api-keys).

## Resources

- [`spork_monitor`](docs/resources/monitor.md) — Manage uptime monitors
- [`spork_alert_channel`](docs/resources/alert_channel.md) — Manage alert channels
- [`spork_status_page`](docs/resources/status_page.md) — Manage status pages

## Data Sources

- [`spork_monitor`](docs/data-sources/monitor.md) — Read a monitor
- [`spork_alert_channel`](docs/data-sources/alert_channel.md) — Read an alert channel
- [`spork_status_page`](docs/data-sources/status_page.md) — Read a status page
- [`spork_status_pages`](docs/data-sources/status_pages.md) — List all status pages

## Prefer the CLI?

Install the [Spork CLI](https://github.com/sporkops/cli) for interactive terminal-based monitoring:

```sh
brew install sporkops/tap/spork && spork login && spork ping add https://yoursite.com
```

## Development

### Building

```shell
go build ./...
```

### Running Acceptance Tests

```shell
export SPORK_API_KEY="your-api-key-here"
TF_ACC=1 go test ./internal/provider/ -v -tags=acceptance -timeout 120m
```

---

**Free to start. No credit card required.** [Sign up at sporkops.com →](https://sporkops.com)

## License

MPL-2.0. See [LICENSE](LICENSE).
