terraform {
  backend "remote" {
    organization = "KiwiSheets"

    workspaces {
      name = "KiwiSheets-Invoicing-Dev"
    }
  }
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

resource "vault_generic_secret" "invoicing" {
  path = "secret/invoicing"

  data_json = jsonencode({
    postgres_password = random_password.postgres_password.result
  })
}

resource "vault_policy" "invoicing" {
  name = "invoicing"

  policy = <<EOT
path "secret/invoicing" {
  capabilities = ["read"]
}
path "secret/data/invoicing" {
  capabilities = ["read"]
}
path "secret/gql-server" {
  capabilities = ["read"]
}
path "secret/data/gql-server" {
  capabilities = ["read"]
}
EOT
}

resource "consul_intention" "postgres" {
  source_name      = "invoicing"
  destination_name = "invoicing-postgres"
  action           = "allow"
}

resource "consul_intention" "rabbit" {
  source_name      = "invoicing"
  destination_name = "rabbitmq"
  action           = "allow"
}

resource "consul_intention" "gql_server" {
  source_name      = "invoicing"
  destination_name = "gql-server"
  action           = "allow"
}

resource "nomad_job" "invoicing" {
  jobspec = templatefile("${path.module}/jobs/job.hcl", {
    datacenter = var.datacenter
    image_tag  = var.image_tag
    instance   = var.instance_count
  })
  detach = false
}
