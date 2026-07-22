variable do_token {
  type = string
}

variable tf_token {
    type = string
}

variable itzg_env {
    type = string
    default = ""
}

variable name {
    type = string
    default = "mc"
}

variable domain {
    type = string
    default = ""
}

variable record {
    type = string
    default = ""
}

variable digitalocean_region {
    type = string
    default = "sgp1"
}

variable digitalocean_volume_size {
    type = string
    default = "4G"
}

