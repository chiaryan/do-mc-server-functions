
resource digitalocean_domain main {
  count = var.create_domain ? 1 : 0
  
  name = var.domain
}

data digitalocean_domain main {
  count = var.create_domain ? 0 : 1

  name = var.domain
}

resource digitalocean_record main {
  domain = var.create_domain ? digitalocean_domain.main[0].id : data.digitalocean_domain.main[0].id
  name = var.record
  value = "0.0.0.0"
  ttl = 30
  type = "A"
}