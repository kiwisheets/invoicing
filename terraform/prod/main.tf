terraform {
  backend "remote" {
    organization = "KiwiSheets"

    workspaces {
      name = "KiwiSheets-GraphQL-Server-Prod"
    }
  }
}

provider "hcloud" {
  token = var.hcloud_token
}

provider "nomad" {}

provider "consul" {
  datacenter = var.datacenter
}

provider "vault" {}

resource "random_password" "postgres_password" {
  length  = 32
  special = false
}

resource "random_password" "jwt_secret" {
  length  = 32
  special = false
}

resource "random_password" "hash_salt" {
  length  = 32
  special = false
}

resource "vault_generic_secret" "gql_server" {
  path = "secret/gql-server"

  data_json = jsonencode({
    postgres_password = random_password.postgres_password.result
    jwt_secret        = random_password.jwt_secret.result
    hash_salt         = random_password.hash_salt.result
  })
}

resource "vault_policy" "gql_server" {
  name = "gql-server"

  policy = <<EOT
path "secret/gql-server" {
  capabilities = ["read"]
}
path "secret/data/gql-server" {
  capabilities = ["read"]
}
EOT
}

data "nomad_plugin" "hcloud_volume" {
  plugin_id        = "hcloud-volume"
  wait_for_healthy = true
}

resource "hcloud_volume" "gql_postgres" {
  name     = "gql-postgres"
  size     = var.postgres_volume_size
  location = "nbg1"
}

resource "nomad_volume" "gql_postgres" {
  depends_on            = [data.nomad_plugin.hcloud_volume]
  type                  = "csi"
  plugin_id             = "hcloud-volume"
  volume_id             = hcloud_volume.gql_postgres.name
  name                  = hcloud_volume.gql_postgres.name
  external_id           = hcloud_volume.gql_postgres.id
  access_mode           = "single-node-writer"
  attachment_mode       = "file-system"
  deregister_on_destroy = true
}

resource "nomad_job" "gql_server" {
  jobspec = templatefile("${path.module}/jobs/gqlserver.hcl", {
    image_tag = var.image_tag
    instance  = var.instance_count
    host      = var.host
    volume_id = nomad_volume.gql_postgres.volume_id
  })
  detach = false
}
