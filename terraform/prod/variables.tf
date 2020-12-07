variable "datacenter" {
  type = string
}

variable "hcloud_token" {
  type        = string
  description = "Hetzner Cloud API Token"
}

variable "image_tag" {
  type        = string
  description = "image version"
}

variable "instance_count" {
  type        = number
  description = "number of server instances to launch"
}

variable "host" {
  type        = string
  description = "API host"
}

variable "postgres_volume_size" {
  type        = number
  description = "postgres volume size"
}
