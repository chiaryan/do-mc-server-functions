variable do_token {
    type = string
}

variable tf_token {
    type = string
}

variable github_app_id {
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
    default = "@"
}

variable create_domain {
    description = "whether to create a domain on digitalocean. set to false if you have a registered domain and want to use it"
    type = bool
    default = true
}

variable digitalocean_region {
    type = string
    default = "sgp1"
}

variable digitalocean_volume_size {
    type = number
    default = 4
}

variable digitalocean_droplet_size {
    type = string
    default = "s-2vcpu-4gb"
}

