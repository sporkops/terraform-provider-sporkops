terraform {
  required_providers {
    spork = {
      source  = "sporkops/sporkops"
      version = "~> 0.6"
    }
  }
}

# Configure the provider using the SPORK_API_KEY environment variable
provider "spork" {}
