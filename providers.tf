terraform {
  required_version = "~> 1.5"
  required_providers {
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = ">= 1.2.0"
    }
    cloudfoundry-community = {
      source  = "cloudfoundry-community/cloudfoundry"
      version = ">= 0.53.1"
    }
  }
}
