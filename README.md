# gitlab-runner-cloudgov
Code for running GitLab CI/CD jobs on cloud.gov

## Status

Currently this just eases the task of deploying the `gitlab-runner` binary and registering it with GitLab. It still uses the `shell` Executor to run jobs, though I've left pointers for adding a `custom` Executor that uses `cf run-task` as the next step.

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
    cf create-service cloud-gov-service-account space-deployer SERVICENAME -t "gitlab-service-account"
    ```

4. Copy `vars.yml-template` to `vars.yml`
    ```
    cp vars.yml-template vars.yml
    ```

5. Edit `vars.yml` and modify the values there as needed. In particular, you must 
    * supply the `ci_server_token` provided when you [configure the runner at the target GitLab URL](https://docs.gitlab.com/ee/tutorials/create_register_first_runner/#create-and-register-a-project-runner)
    * supply the `service_account_instance` name that you used when you created the service instance in a previous step

6. Deploy the GitLab runner
    ```
    cf push --vars-file vars.yml
    ```
7. Check to see that the runner has registered itself in GitLab

## TODO

- Add a `gitlab-worker` app
- Write a [custom `gitlab-runner` Executor](https://docs.gitlab.com/runner/executors/custom.html) that uses `cf ssh gitlab-worker $@` (refer to [the libvirt example](https://docs.gitlab.com/runner/executors/custom_examples/libvirt.html))
- [Configure a bound S3 bucket as the cache](https://docs.gitlab.com/runner/configuration/advanced-configuration.html#the-runnerscaches3-section)
- Tests!
- Documentation

## Design Decisions

### Use environment variables to register gitlab-runner

Recent versions of `gitlab-runner` expose almost all initial configuration
variables for the `register` subcommand as environment variables. This allows
us to do almost all configuration in `manifest.yml` and skip modifying
command line options in `runner/.profile` or having a .toml add on.
