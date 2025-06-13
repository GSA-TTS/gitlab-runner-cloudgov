locals {
  allowlist_map = {
    cloud_gov = ["*.fr.cloud.gov"]
    ruby = [
      "*.rubygems.org"
    ]
    terraform = [
      "dl-cdn.alpinelinux.org", # allow installing extra dependencies in terraform images, like `zip`
      "*.releases.hashicorp.com",
      "registry.terraform.io",
      "objects.githubusercontent.com",
      "release-assets.githubusercontent.com",
      "sts.us-gov-west-1.amazonaws.com"
    ]
    node = [
      "deb.debian.org", # allow updating system in concert with node deb install
      "deb.nodesource.com",
      "registry.npmjs.org",
      "*.yarnpkg.com"
    ]
    oscal = [
      "raw.githubusercontent.com"
    ]
    containers = [
      "*.docker.io",
      "*.docker.com",
      "*.ghcr.io",
      "*.gcr.io",
      "registry.${var.ci_server_url}"
    ]
    debian = ["deb.debian.org"]
    ubuntu = ["*.ubuntu.com"]
    alpine = ["dl-cdn.alpinelinux.org"]
    fedora = ["*.fedoraproject.org"]
  }
}
