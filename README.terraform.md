# gitlab-runner-cloudgov terraform

Terraform for running GitLab CI/CD jobs on cloud.gov or another CloudFoundry based PaaS.

* [Deploying](#deploying)
* [Troubleshooting](#troubleshooting)
* [Design Decisions](#design-decisions)

## Deploying

Include this repository as a module in your terraform setup to deploy a gitlab-runner-cloudgov system.

```terraform
module "gitlab-runner" {
    source = "github.com/gsa-tts/gitlab-runner-cloudgov?ref=main" # TODO: releases via tags

    cf_space_prefix  = "SPACEPREFIX"
    ci_server_token  = "TOKEN"
    docker_hub_user  = "docker-username"
    docker_hub_token = "docker-personal-access-token"
    worker_egress_allowlist = [
        "hostname.to.allow", # explanation of why it's allowed
        "other.hostname.to.allow" # explanation of why it's allowed
    ]
}
```

### Local terraform use / sandbox deploy

For local development, there is a `sandbox-deploy` module set up to deploy the root module that will use local state storage.

1. Log in to cloud.gov and select your ORGNAME when prompted
    ```
    cf login -a api.fr.cloud.gov --sso
    ```

1. Create a management space, if it does not already exist.
    ```
    cf create-space SPACEPREFIX-mgmt
    ```

1. Switch working directory
    ```
    cd sandbox-deploy
    ```

1. Create a [cloud.gov service account](https://cloud.gov/docs/services/cloud-gov-service-account/) with the `OrgManager` permission
    ```
    ./create_service_account -s SPACEPREFIX-mgmt -u glr-local-deploy -m > secrets.auto.tfvars
    ```

1. Copy `vars.tfvars-example` to `vars.auto.tfvars`.
    ```
    cp vars.tfvars-example vars.auto.tfvars
    ```

1. Edit `vars.auto.tfvars` and modify the values there as needed. In particular, you must:
    * for sandbox/developer deployments, set `cf_space_prefix` to the same as `SPACEPREFIX` used for the management space
    * supply the `ci_server_token` provided when you [configure the runner at the target GitLab URL](https://docs.gitlab.com/ee/tutorials/create_register_first_runner/#create-and-register-a-project-runner)
    * supply a docker hub username and personal access token to avoid rate limiting
    * set `developer_emails` to whoever might need to debug this deployment
    * set `worker_egress_allowlist` to the package hosts needed for your supported programming languages

1. Run `terraform init`

1. Run `terraform validate`

1. Run `terraform plan` and double check that the changes are what is expected.

1. Apply your changes with `terraform apply`

1. Check to see that the runner has registered itself in GitLab under your project repository under Settings -> CI/CD -> Runners (Expand)

At this point the runner should be available to run jobs. See [Use GitLab - Use CI/CD to build your application - Getting started](https://docs.gitlab.com/ee/ci/)
for much more on GitLab CI/CD and runners.

## Troubleshooting

### Viewing manager instance logs

Problems with runner registration often requiring viewing its logs.

~~~
cf target -s SPACEPREFIX-manager
cf logs --recent devtools-runner-manager
~~~

### Dependency installs are not working, dependencies cannot be downloaded.

The manager and workers run in [restricted-egress](https://cloud.gov/docs/management/space-egress/) spaces. There are two places to edit in order to allow traffic.

1. If the runner-manager cannot download something, or the runner-workers are failing during the `prepare.sh` steps then the `local.devtools_egress_allowlist` in `main.tf` should be updated
1. If the runner-workers cannot download a dependency required because of the programming language in use by the project, then it should likely be added to the `var.worker_egress_allowlist` in `vars.auto.tfvars` or your module include.

It is also possible that additional configuration is required for the package manager in question to direct traffic over the proxy.

## TODO

## Design Decisions

### Use environment variables to register gitlab-runner

Recent versions of `gitlab-runner` expose almost all initial configuration
variables for the `register` subcommand as environment variables. This allows
us to do almost all configuration in `manifest.yml` and skip modifying
command line options in `runner/.profile` or having a .toml add on.
