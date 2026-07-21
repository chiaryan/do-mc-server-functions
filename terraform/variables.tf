variable dotoken {
  type = string
  description = "DigitalOcean Token (starts with dop_....)"
}

variable region {
  type = string
  default = "sgp1"
}

variable itzg_env { type = string }

variable stop_function_address { type = string }

variable stop_function_token { type = string }

variable domain { type = string }

variable size {
  type = string
  default = "s-2vcpu-4gb"
}

variable record {
  type = string
  default = ""
}

variable volume_name {
  type = string
  default = "mc-data"
}

variable auto_destroy {
  type = bool
  default = true
}

variable ssh_key {
  type = string
  default = ""
}