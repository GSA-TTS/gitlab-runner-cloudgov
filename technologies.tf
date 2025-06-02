locals {
  allowlist_map = {
    ruby = [
      "*.rubygems.org"
    ]
    terraform = [
      "*.releases.hashicorp.com",
      "registry.terraform.io",
      "objects.githubusercontent.com",
      "sts.us-gov-west-1.amazonaws.com"
    ]
    node = [
      "deb.nodesource.com",
      "registry.npmjs.org",
      "*.yarnpkg.com"
    ]
    oscal = [
      "raw.githubusercontent.com"
    ]
  }
}
