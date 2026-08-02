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
      dynamic env {
        for_each = local.envs

        content {
          key = env.key
          value = env.value
        }
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