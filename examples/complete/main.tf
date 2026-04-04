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
