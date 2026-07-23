terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
    tfe = {
      source = "hashicorp/tfe"
      version = "~> 0.79.0"
    }
    tls = {
      source = "hashicorp/tls"
    }
  }
}


provider digitalocean {
  token = var.do_token
}

provider tfe {
  token = var.tf_token
}

resource "random_password" "main" {
  length = 32
}

resource digitalocean_domain main {
  count = var.create_domain ? 1 : 0
  
  name = var.domain
}

data digitalocean_domain main {
  count = var.create_domain ? 0 : 1

  name = var.domain
}

resource digitalocean_app main {
  spec {
    name = "${var.name}-app"
    region = var.digitalocean_region

    function {
      name = var.name
      github {
        repo = "chiaryan/do-mc-server-functions"
        branch = "master"
      }
    }
    env {
      key = "TFE_TOKEN"
      value = var.tf_token
    }
    env {
      key = "WORKSPACE_ID"
      value = tfe_workspace.main.id
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
      key = "STOP_ADDRESS"
      value = "$${_self.FUNCTION_URL}"
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
      key = "INSTANCE_VOLUME_NAME"
      value = digitalocean_ssh_key.main.name
    }
    env {
      key = "STOP_ADDRESS_PASSWORD_HASH"
      value = random_password.main.bcrypt_hash
    }
    env {
      key = "STOP_ADDRESS_PASSWORD"
      value = random_password.main.result
    }
  }
}


resource digitalocean_volume main {
  name = "${var.name}-vol"
  region = var.digitalocean_region
  size = var.digitalocean_volume_size
  initial_filesystem_type = "ext4"
}