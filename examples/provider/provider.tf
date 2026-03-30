terraform {
  required_providers {
    spork = {
      source = "sporkops/sporkops"
    }
  }
}

# Configure the Spork provider.
# The API key can also be set via the SPORK_API_KEY environment variable.
provider "spork" {
  api_key = var.spork_api_key
}

variable "spork_api_key" {
  type      = string
  sensitive = true
}
