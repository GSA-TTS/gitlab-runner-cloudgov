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

3. Copy `vars.yml-template` to `vars.yml`
    ```
    cp vars.yml-template vars.yml
    ```

4. Edit `vars.yml` and modify the values there as needed. In particular, you must supply an `authentication_token` provided by the target GitLab URL.

5. Deploy the GitLab runner
    ```
    cf push --vars-file vars.yml
    ```
6. Check to see that the runner has registered itself in GitLab

## TODO

- Add a `gitlab-worker` app
- Write a [custom `gitlab-runner` Executor](https://docs.gitlab.com/runner/executors/custom.html) that uses `cf run-task gitlab-runner -w -c $@` (refer to [the libvirt example](https://docs.gitlab.com/runner/executors/custom_examples/libvirt.html))
