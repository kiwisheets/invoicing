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
          "secrets/hash-salt.secret:/run/secrets/hash-salt.secret",
          "secrets/rabbitmq-dsn.secret:/run/secrets/rabbitmq-dsn.secret"
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
        REDIS_ADDRESS = "$${NOMAD_UPSTREAM_ADDR_invoicing-redis}"
        HASH_SALT_FILE = "/run/secrets/hash-salt.secret"
        HASH_MIN_LENGTH = 10
        RABBITMQ_DSN_FILE = "/run/secrets/rabbitmq-dsn.secret"
        GQL_SERVER_URL = "http://localhost:8000/graphql"
      }

      template {
        data = "{{with secret \"secret/data/invoicing\"}}{{.Data.data.postgres_password}}{{end}}"
        destination = "secrets/db-password.secret"
      }

      template {
        data = "{{with secret \"secret/data/gql-server\"}}{{.Data.data.hash_salt}}{{end}}"
        destination = "secrets/hash-salt.secret"
      }

      template {
        data = "amqp://admin:{{with secret \"secret/data/rabbitmq\"}}{{.Data.data.rabbitmq_password}}{{end}}@localhost:5672"
        destination = "secrets/rabbitmq-dsn.secret"
      }

      vault {
        policies = [
          "invoicing", 
          "rabbitmq"
        ]
      }

      resources {
        cpu    = 64
        memory = 64
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
              local_bind_port  = 5432
            }
            upstreams {
              destination_name = "invoicing-redis"
              local_bind_port = 6379
            }
            upstreams {
              destination_name = "rabbitmq"
              local_bind_port  = 5672
            }
            upstreams {
              destination_name = "gql-server"
              local_bind_port  = 8000
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
        data = "{{with secret \"secret/data/invoicing\"}}{{.Data.data.postgres_password}}{{end}}"
        destination = "secrets/db-password.secret"
      }

      vault {
        policies = ["invoicing"]
      }

      resources {
        cpu    = 50
        memory = 64
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

  group "redis" {
    count = 1

    task "redis" {
      driver = "docker"

      config {
        image = "redis:latest"
      }

      resources {
        cpu    = 50
        memory = 64
      }
    }

    network {
      mode = "bridge"
    }

    service {
      name = "invoicing-redis"
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
