terraform {
  required_providers {
    spork = {
      source  = "sporkops/sporkops"
      version = "~> 1.0"
    }
  }
}

# Configure the provider using the SPORK_API_KEY environment variable
provider "spork" {}
