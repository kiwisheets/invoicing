terraform {
  required_providers {
    nomad = {
      source = "hashicorp/nomad"
    }
    consul = {
      source  = "hashicorp/consul"
      version = "2.10.1"
    }
    vault = {
      source = "hashicorp/vault"
    }
  }
  required_version = ">= 0.13"
}
