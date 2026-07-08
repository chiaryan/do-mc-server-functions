terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

variable dotoken {
  type = string
  description = "DigitalOcean Token (starts with dop_....)"
}

variable region {
  type = string
  default = "sgp1"
}

variable itzg_env {
  type = string
}

variable stop_function_address {
  type = string
}

variable domain {
  type = string
}

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

variable ssh_keys {
  type = list(string)
  default = []
}

provider "digitalocean" {
  token = var.dotoken
}

data "digitalocean_volume" "main" {
  name = var.volume_name
}

locals {
  cloud_config = yamlencode({
    #cloud-config
    mounts = [
      [ "/dev/disk/by-id/scsi-0DO_Volume_${var.volume_name}", "/mnt/data", "ext4", "defaults,nofail,discard", "0", "0"]
    ]
    runcmd = concat(
      ["docker run -v /mnt/data:/data -i -p 25565:25565 --env-file .env itzg/minecraft-server"],
      var.auto_destroy ? ["curl ${var.stop_function_address}"] : []
    )
    write_files = [
      {
        content = var.itzg_env
        path = "/.env"
      }
    ]
  })
}

resource "digitalocean_droplet" "main" {
  image = "docker-20-04"
  name = "mc-server"
  region = var.region
  size = var.size

  user_data = "#cloud-config\n${local.cloud_config}"
  volume_ids = [ data.digitalocean_volume.main.id ]
  monitoring = true
  ssh_keys = var.ssh_keys
}

data "digitalocean_domain" "main" {
  name = var.domain
}

resource "digitalocean_record" "main" {
  domain = data.digitalocean_domain.main.id
  name = var.record
  value = digitalocean_droplet.main.ipv4_address
  ttl = 30
  type = "A"
}