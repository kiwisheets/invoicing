terraform {
  required_providers {
    nomad = {
      source = "hashicorp/nomad"
    }
    consul = {
      source = "hashicorp/consul"
    }
    vault = {
      source = "hashicorp/vault"
    }
    hcloud = {
      source = "hetznercloud/hcloud"
    }
  }
  required_version = ">= 0.13"
}
