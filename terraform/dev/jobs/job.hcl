job "invoicing" {
  datacenters = ["${datacenter}"]

  group "invoicing" {
    count = 1

    task "invoicing" {
      driver = "docker"

      config {
        image = "kiwisheets/invoicing:${image_tag}"

        volumes = [
          "secrets/db-password.secret:/run/secrets/db-password.secret",
          "secrets/jwt-secret-key.secret:/run/secrets/jwt-secret-key.secret",
          "secrets/hash-salt.secret:/run/secrets/hash-salt.secret"
        ]
      }

      env {
        APP_VERSION = "0.0.0"
        API_PATH = "/graphql"
        PORT = 3000
        ENVIRONMENT = "production"
        POSTGRES_HOST = "$${NOMAD_UPSTREAM_IP_invoicing-postgres}"
        POSTGRES_PORT = "$${NOMAD_UPSTREAM_PORT_invoicing-postgres}"
        POSTGRES_DB = "invoicing"
        POSTGRES_USER = "invoicing"
        POSTGRES_PASSWORD_FILE = "/run/secrets/db-password.secret"
        POSTGRES_MAX_CONNECTIONS = 20
        HASH_SALT_FILE = "/run/secrets/hash-salt.secret"
        HASH_MIN_LENGTH = 10
      }

      template {
        data = <<EOF
{{with secret "secret/data/invoicing"}}{{.Data.data.postgres_password}}{{end}}
        EOF
        destination = "secrets/db-password.secret"
      }

      template {
        data = <<EOF
{{with secret "secret/data/gql-server"}}{{.Data.data.hash_salt}}{{end}}
        EOF
        destination = "secrets/hash-salt.secret"
      }

      vault {
        policies = ["invoicing"]
      }

      resources {
        cpu    = 256
        memory = 256
      }
    }

    network {
      mode = "bridge"
      port "health" {}
    }

    service {
      name = "invoicing"
      port = 3000

      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "invoicing-postgres"
              local_bind_port = 5432
            }
            expose {
              path {
                path           = "/health"
                protocol        = "http"
                local_path_port = 3000
                listener_port   = "health"
              }
            }
          }
        }

        sidecar_task {
          resources {
            cpu    = 20
            memory = 32
          }
        }
      }

      check {
        type     = "http"
        path     = "/health"
        port     = "health"
        interval = "2s"
        timeout  = "2s"
      }
    }
  }

  group "postgres" {
    count = 1

    task "postgres" {
      driver = "docker"

      config {
        image = "postgres:latest"

        volumes = [
          "secrets/db-password.secret:/run/secrets/db-password.secret"
        ]
      }

      env {
        PGDATA = "/var/lib/postgresql/data/db"
        POSTGRES_DB = "invoicing"
        POSTGRES_USER = "invoicing"
        POSTGRES_PASSWORD_FILE = "/run/secrets/db-password.secret"
      }

      template {
        data = <<EOF
{{with secret "secret/data/invoicing"}}{{.Data.data.postgres_password}}{{end}}
        EOF
        destination = "secrets/db-password.secret"
      }

      vault {
        policies = ["invoicing"]
      }
    }
    
    network {
      mode = "bridge"
    }

    service {
       name = "invoicing-postgres"
       port = "5432"

      connect {
        sidecar_service {}

        sidecar_task {
          resources {
            cpu    = 20
            memory = 32
          }
        }
      }
    }
  }
}
