# Terraform Provider for Spork

The Spork Terraform provider allows you to manage [Spork](https://sporkops.com) uptime monitors and alert channels as infrastructure-as-code.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- [Go](https://go.dev/dl/) >= 1.22 (for building from source)

## Usage

```hcl
terraform {
  required_providers {
    spork = {
      source  = "sporkops/sporkops"
      version = "~> 1.0"
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

## Registry

Manage published provider versions in the [Terraform Registry](https://app.terraform.io/app/sporkops/registry/public-namespaces/sporkops/providers).

## Development

### Building

```shell
go build ./...
```

### Running Acceptance Tests

Acceptance tests require a valid API key and create real resources:

```shell
export SPORK_API_KEY="your-api-key-here"
TF_ACC=1 go test ./internal/provider/ -v -tags=acceptance -timeout 120m
```

### Installing Locally

```shell
make install
```

This builds the provider and installs it to `~/.terraform.d/plugins/` for local development.

## License

MPL-2.0. See [LICENSE](LICENSE).