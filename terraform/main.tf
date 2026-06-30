terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = "TODO"
}

resource "digitalocean_droplet" "main" {
  image = "ubuntu-22-04-x64"
  name = "www-1"
  region = "nyc2"
  size = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
  user_data = <<EOF
    #cloud-config
    package_update: true
    package_upgrade: true
    packages:
      - nginx
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
  EOF
}

resource "digitalocean_record" "main" {
  domain = "chiaryan.xyz"
  name = "exmc.chiaryan.xyz"
  value = digitalocean_droplet.main.ipv4_address
  type = "A"
}