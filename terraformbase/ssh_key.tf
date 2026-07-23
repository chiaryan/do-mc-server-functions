resource "tls_private_key" "name" {
  algorithm = "ED25519"
  rsa_bits = 2048
}

resource digitalocean_ssh_key main {
  name = "${var.name}-key"
  public_key = tls_private_key.name.public_key_openssh
}

resource "local_file" "ssh_key" {
  content = tls_private_key.name.private_key_openssh
  filename = "${var.name}-key.pem"
}