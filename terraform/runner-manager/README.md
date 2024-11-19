# gitlab-runner-cloudgov terraform
Terraform for running GitLab CI/CD jobs on cloud.gov or another CloudFoundry based PaaS.

* [Deploying](#deploying)
* [Troubleshooting](#troubleshooting)
* [Design Decisions](#design-decisions)

## Deploying

1. Log in to cloud.gov
    ```
    cf login -a api.fr.cloud.gov --sso
    ```

2. Target the org and space for deployment
    ```
    cf target -o ORGNAME -s SPACENAME
    ```
    For example:
    ```
    cf target -o sandbox-gsa -s bret.mogilefsky
    ```

3. Create a [cloud.gov service account](https://cloud.gov/docs/services/cloud-gov-service-account/), tagged with `gitlab-service-account`
    ```
    cf create-service cloud-gov-service-account space-deployer SERVICE_ACCOUNT_INSTANCE -t "gitlab-service-account"
    ```

4. Copy `vars.tfvars` to `vars.auto.tfvars`. 
    ```
    cp vars.tfvars vars.auto.tfvars
    ```

5. Edit `vars.auto.tfvars` and modify the values there as needed. In particular, you must 
    * supply the `ci_server_token` provided when you [configure the runner at the target GitLab URL](https://docs.gitlab.com/ee/tutorials/create_register_first_runner/#create-and-register-a-project-runner)
    * supply the `service_account_instance` name that you used when you created the service instance in step 3
    * supply the `object_store_instance` name that you would like to use.

6. In the /terraform/runner directory run terraform init
    ```
    terraform init
    ```

7. In the /terraform/runner directory, validate your terraform
    ```
    terraform validate
    ```

8. In the /terraform/runner directory, run terraform non-destructively
    ```
    terraform plan
    ```   

9. In the /terraform/runner directory, apply your terraform
    ```
    terraform apply
    ```   

7. Check to see that the runner has registered itself in GitLab under your project
   repository under Settings -> CI/CD -> Runners (Expand)

At this point the runner should be available to run jobs. See [Use GitLab - Use CI/CD to build your application - Getting started](https://docs.gitlab.com/ee/ci/)
for much more on GitLab CI/CD and runners.

## Troubleshooting

### Viewing manager instance logs

Problems with runner registration often requiring viewing it's logs.

~~~
cf logs --recent RUNNER-NAME
~~~

### "Request error: Get https://API-URL/v2/info: dial tcp X.X.X.X:443: connect: connection refused"

The GitLab Runner manager needs to contact the CloudFoundry API to schedule
runner applications. This indicates your CloudFoundry space security group may
be too restrictive or not set.

For a production deployment you should use tightly controlled egress filtering,
ideally with a name based proxy.

Test Only - For a basic test environment with no privileged access you can use
the following to apply a loose egress security group policy on cloud.gov:

~~~
cf bind-security-group public_networks_egress ORG_NAME --space SPACE_NAME
~~~

## TODO

## Design Decisions

### Use environment variables to register gitlab-runner

Recent versions of `gitlab-runner` expose almost all initial configuration
variables for the `register` subcommand as environment variables. This allows
us to do almost all configuration in `manifest.yml` and skip modifying
command line options in `runner/.profile` or having a .toml add on.
