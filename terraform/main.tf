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

variable domain {
  type = string
}

variable record {
  type = string
  default = ""
}

variable volume_name {
  type = string
  default = "mc-data"
}

provider "digitalocean" {
  token = var.dotoken
}

data "digitalocean_volume" "main" {
  name = var.volume_name
}

resource "digitalocean_droplet" "main" {
  image = "docker-20-04"
  name = "mc-server"
  region = var.region
  size = "s-2vcpu-4gb"

  # user_data = templatefile("cloud-config.yaml", {
  #   DATA_VOL = var.volume_name
  #   ITZG_ENV = var.itzg_env
  # })
  user_data = yamlencode({
    #cloud-config
    mounts = [
      [ "/dev/disk/by-id/scsi-0DO_Volume_${var.volume_name}", "/mnt/data", "ext4", "defaults,nofail,discard", "0", "0"]
    ]
    runcmd = [
      "docker run -v /mnt/data:/data -p 25565:25565 --env-file .env itzg/minecraft-server"
    ]
    write_files = [
      {
        content = var.itzg_env
        path = "/.env"
      }
    ]
  }
  )
  volume_ids = [ data.digitalocean_volume.main.id ]
  monitoring = true
  ssh_keys = ["63:fa:78:dd:49:02:63:bd:f1:6c:ad:ed:fd:78:03:d1"]
}

data "digitalocean_domain" "main" {
  name = var.domain
}

resource "digitalocean_record" "main" {
  domain = data.digitalocean_domain.main.id
  name = var.record
  value = digitalocean_droplet.main.ipv4_address
  type = "A"
}