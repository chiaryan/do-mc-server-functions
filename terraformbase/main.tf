terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    tls = {
      source = "hashicorp/tls"
    }
  }
}


provider digitalocean {
  token = var.do_token
}

resource "random_password" "main" {
  length = 32
}

variable github_repo {
  type = string
  default = "chiaryan/do-mc-server-functions"
}

variable github_branch {
  type = string
  default = "master"
}

resource digitalocean_app main {
  spec {
    name = "${var.name}-app"
    region = var.digitalocean_region

    function {
      name = "functions"
      github {
        repo = var.github_repo
        branch = var.github_branch
      }
      env {
        key = "SERVER_DOMAIN"
        value = "${var.record}.${var.domain}"
      }
      env {
        key = "DO_TOKEN"
        value = var.do_token
      }
      env {
        key = "ITZG_ENV"
        value = var.itzg_env
      }
      env {
        key = "RECORD"
        value = var.record
      }
      env {
        key = "DOMAIN"
        value = var.domain
      }
      env {
        key = "INSTANCE_SSH_KEY"
        value = digitalocean_ssh_key.main.fingerprint
      }
      env {
        key = "INSTANCE_SIZE"
        value = var.digitalocean_droplet_size
      }
      env {
        key = "INSTANCE_NAME"
        value = "${var.name}-minecraft"
      }
      env {
        key = "INSTANCE_VOLUME_NAME"
        value = "${var.name}-vol"
      }
      env {
        key = "INSTANCE_VOLUME_ID"
        value = "${digitalocean_volume.main.id}"
      }
      env {
        key = "FUNCTIONS_PASSWORD_HASH"
        value = random_password.main.bcrypt_hash
      }
      env {
        key = "FUNCTIONS_PASSWORD"
        value = random_password.main.result
      }
      env {
        key = "SSH_KEY"
        value = tls_private_key.name.public_key_fingerprint_md5
      }
    }
  }
}


resource digitalocean_volume main {
  name = "${var.name}-vol"
  region = var.digitalocean_region
  size = var.digitalocean_volume_size
  initial_filesystem_type = "ext4"
}