locals {
  allowlist_map = {
    ruby = [
      "*.rubygems.org"
    ]
    terraform = [
      "*.releases.hashicorp.com",
      "registry.terraform.io",
      "objects.githubusercontent.com"
    ]
    node = [
      "deb.nodesource.com",
      "registry.npmjs.org",
      "registry.yarnpkg.com"
    ]
  }
}
