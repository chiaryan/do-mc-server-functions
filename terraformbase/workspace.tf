resource tfe_organization main {
  name = "${var.name}-organization"
  email = "mail@example.com"
}

resource tfe_workspace main {
  name = "${var.name}-workspace"
  organization = tfe_organization.main.id
  vcs_repo {
    identifier = "chiaryan/do-mc-server-functions"
    branch = "master"
    github_app_installation_id = var.github_app_id
  }
  working_directory = "terraform"
}

