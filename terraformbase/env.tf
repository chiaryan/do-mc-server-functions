locals {
  envs = {
    SERVER_DOMAIN = "${var.record}.${var.domain}"
    RECORD_ID = digitalocean_record.main.id
    DO_TOKEN = var.do_token
    ITZG_ENV = var.itzg_env
    INSTANCE_SSH_KEY = digitalocean_ssh_key.main.fingerprint
    INSTANCE_SIZE = var.digitalocean_droplet_size
    INSTANCE_NAME = "${var.name}-minecraft"
    INSTANCE_REGION = var.digitalocean_region
    INSTANCE_VOLUME_NAME = "${var.name}-vol"
    INSTANCE_VOLUME_ID = "${digitalocean_volume.main.id}"
    FUNCTIONS_PASSWORD_HASH = random_password.main.bcrypt_hash
    FUNCTIONS_URL = "$${_self.PUBLIC_URL}"
    FUNCTIONS_PASSWORD = random_password.main.result
    AUTO_DESTROY = var.auto_destroy
  }
}