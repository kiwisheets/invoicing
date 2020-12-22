variable "datacenter" {
  type = string
}

variable "image_tag" {
  type        = string
  description = "image version"
}

variable "instance_count" {
  type        = number
  description = "number of server instances to launch"
}
