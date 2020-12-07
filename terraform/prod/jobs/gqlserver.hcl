job "gql-server" {
  datacenters = ["${datacenter}"]

  group "gql-server" {
    count = 1

    task "gql-server" {
      driver = "docker"

      config {
        image = "kiwisheets/gql-server:${image_tag}"

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
        POSTGRES_HOST = "$${NOMAD_UPSTREAM_IP_gql-postgres}"
        POSTGRES_PORT = "$${NOMAD_UPSTREAM_PORT_gql-postgres}"
        POSTGRES_DB = "kiwisheets"
        POSTGRES_USER = "kiwisheets"
        POSTGRES_PASSWORD_FILE = "/run/secrets/db-password.secret"
        POSTGRES_MAX_CONNECTIONS = 20
        REDIS_ADDRESS = "$${NOMAD_UPSTREAM_ADDR_gql-redis}"
        JWT_SECRET_KEY_FILE = "/run/secrets/jwt-secret-key.secret"
        HASH_SALT_FILE = "/run/secrets/hash-salt.secret"
        HASH_MIN_LENGTH = 10
      }

      template {
        data = <<EOF
{{with secret "secret/data/gql-server"}}{{.Data.data.postgres_password}}{{end}}
        EOF
        destination = "secrets/db-password.secret"
      }

      template {
        data = <<EOF
{{with secret "secret/data/gql-server"}}{{.Data.data.jwt_secret}}{{end}}
        EOF
        destination = "secrets/jwt-secret-key.secret"
      }

      template {
        data = <<EOF
{{with secret "secret/data/gql-server"}}{{.Data.data.hash_salt}}{{end}}
        EOF
        destination = "secrets/hash-salt.secret"
      }

      vault {
        policies = ["gql-server"]
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
      name = "gql-server"
      port = 3000

      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "gql-postgres"
              local_bind_port = 5432
            }
            upstreams {
              destination_name = "gql-redis"
              local_bind_port = 6379
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

    volume "gql-postgres" {
      type      = "csi"
      read_only = false
      source    = "${volume_id}"
    }

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
        POSTGRES_DB = "kiwisheets"
        POSTGRES_USER = "kiwisheets"
        POSTGRES_PASSWORD_FILE = "/run/secrets/db-password.secret"
      }

      volume_mount {
        volume      = "gql-postgres"
        destination = "/var/lib/postgresql/data"
        read_only   = false
      }

      template {
        data = <<EOF
{{with secret "secret/data/gql-server"}}{{.Data.data.postgres_password}}{{end}}
        EOF
        destination = "secrets/db-password.secret"
      }

      vault {
        policies = ["gql-server"]
      }
    }
    
    network {
      mode = "bridge"
    }

    service {
       name = "gql-postgres"
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

  group "redis" {
    count = 1

    task "redis" {
      driver = "docker"

      config {
        image = "redis:latest"
      }
    }

    network {
      mode = "bridge"
    }

    service {
       name = "gql-redis"
       port = "6379"

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
