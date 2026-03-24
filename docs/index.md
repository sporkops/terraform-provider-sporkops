---
page_title: "Provider: Spork"
description: |-
  The Spork provider is used to manage uptime monitors and alert channels.
---

# Spork Provider

The Spork provider allows you to manage [Spork](https://sporkops.com) uptime monitors and alert channels using Terraform.

## Authentication

The provider requires an API key to authenticate with the Spork API. You can provide the key in two ways:

1. **Environment variable** (recommended):

```shell
export SPORK_API_KEY="your-api-key-here"
```

2. **Provider configuration**:

```hcl
provider "spork" {
  api_key = "your-api-key-here"
}
```

### Generating an API Key

Generate an API key from the Spork dashboard at [sporkops.com/settings/api-keys](https://sporkops.com/settings/api-keys).

## Example Usage

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

## Schema

### Optional

- `api_key` (String, Sensitive) — Spork API key. Can also be set via the `SPORK_API_KEY` environment variable.
