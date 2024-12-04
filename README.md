# gitlab-runner-cloudgov
Code for running GitLab CI/CD jobs on cloud.gov or another CloudFoundry based
PaaS.


* [Differences from other GitLab Runner executor types](#differences-from-other-gitlab-runner-executor-types)
* [How it works](#how-it-works)
* [Deploying](#deploying-and-troubleshooting)
* [Design Decisions](#design-decisions)


## Differences from other GitLab Runner executor types

The goal of this runner executor is to support most CI/CD use cases without
needing to modify your `.gitlab-ci.yml`. Please note the current limitiations
and differences in comparison to the Docker executor:

* __No shared filesystem or direct exec capability__ - Some
  executors can share filesystems between containers and directly execute processes
  in containers. CloudFoundry does not support sharing filesystems and only supports
  use of SSH to execute into a running container. This runner attempts to transparently
  work around these limitations where possible. Your CI job may require significant
  modification if it relies on either of these features.
* __Use CI_SERVICE_alias for service DNS names__ - Ephemeral networks like the Docker networks
  used for the Docker executor are not available in CloudFoundry. This means
  that each service you create lives in a common internal DNS namespace with
  other services in the same CloudFoundry space. The cloud.gov runner populates
  a `CI_SERCVCE_<alias>` variable for each service defined in a job. Here is an
  example snippet of GitLab CI Yaml showing the definition of a service and how
  the job steps can then connect to that service:
  ~~~yaml
  # Start a HTTP "echo" service and then send a request to it
  echo-test-job:
    stage: test
    image: ubuntu:24.04
  services:
  - name: http-https-echo:latest
    # Note "echo" is the alias name, so the CI_SERVICE_alias variable key name
    # for this service is CI_SERVICE_echo
    alias: echo
  script:
  # Using the CI_SERVICE_alias to provide the FQDN of the service
  - curl http://${CI_SERVICE_echo}:8080
  ~~~

## How it works

This is a custom executor borrowing ideas from https://docs.gitlab.com/runner/executors/custom.html.

It runs a "manager" GitLab Runner instance responsible for listening for new
jobs. Each job is run on a new application instance in CloudFoundry using the
specified OCI image.

![Fig 1 - Components](doc/gitlab-runner-cf-driver-components.png)
__[source](doc/gitlab-runner-cf-driver-components.pu)__

The relevant components of the solution are shown above with two running stages.
* "GitLab Runner - Manager" - This is the persistent application defined in this
  very repository. It registers with the GitLab instance and waits for further
  instructions.
* "GitLab Runner - Worker" - The manager starts worker application instances
  using the container image you specify and runs job steps via CloudFoundry
  `cf ssh` calls into the running containers.
* "GitLab Runner - Service" - Optional service instances can be started for use
  by the Worker instances. These run as separate apps using the container
  image you specify.

![Fig 2 - Job sequence overview](doc/gitlab-runner-cf-driver-sequence.png)
__[source](doc/gitlab-runner-cf-sequence-overview.pu)__

The above sequence diagram illustrates the series of events in a given job run.
Optional service lifecycles are not shown, but run as part of the `prepare` phase.

The runner manager registration and other flow details are shown
in [Runner Execution Flow](https://gitlab.com/gitlab-org/gitlab-runner/-/tree/main/docs?ref_type=heads#runner-execution-flow).

## Prerequisites

* **GitLab access and ownership of one or more projects/repos** - The runners
  need a GitLab service to get their instructions from. This is a good time to
  familiarize yourself with [registering runners](https://docs.gitlab.com/runner/register/)
  and generating runner tokens.

* **A cloud.gov organization** - You need a cloud.gov organization to deploy your runner(s) into. See
  [Get access to cloud.gov](https://cloud.gov/docs/getting-started/accounts/)
  for details.

* **A cloud.gov space to host the deployment service account** - Runners should only be deployed
  into their own dedicated spaces. The terraform in terraform/runner-manager takes care of this.

* **Network access from your runner space to required Internet destinations** -
  Runners need to reach your GitLab server via HTTPS. They also need to be able
  to reach package repositories and other resources used by your CI/CD jobs.
  The specific destinations vary. Update the proxy rules you use with the `local.devtools_egress_allowlist` and `var.worker_egress_allowlist` configuration values

## Deploying and Troubleshooting

See [the terraform README](terraform/runner-manager/README.md) for deployment instructions and troubleshooting tips.

## TODO

- Make it possible to run DAST/a11y/other tests requiring a running copy of the
  application and a separate scanning process.
- Set custom executor timeouts based on measurements.
- Add linting and tests.
- Improve documentation.
- Add support for RedHat based images in `prepare.sh`.
- Add version pinning and support for other architectures to `gitlab-runner-helper`
  installation in `prepare.sh`.

## Design Decisions

### Use environment variables to register gitlab-runner

Recent versions of `gitlab-runner` expose almost all initial configuration
variables for the `register` subcommand as environment variables. This allows
us to do almost all configuration in `manifest.yml` and skip modifying
command line options in `runner/.profile` or having a .toml add on.

## Compliance Documentation

Security controls and other compliance documents for the DevTools SaaS system live at https://github.com/GSA-TTS/devtools-compliance and also as a git submodule under `/doc/compliance`

From this repo, run `bin/trestle` to start a trestle CLI with helper scripts for maintaining and updating the documents.

See the [docker-trestle README](https://github.com/gsa-tts/docker-trestle) for help with the individual scripts.

### Git Submodule Commands

See git's [submodule documentation](https://git-scm.com/book/en/v2/Git-Tools-Submodules)
for more information on tracking changes to the compliance files.

#### Cloning this project

`git clone --recurse-submodules <<REPO_ADDRESS>>`

#### Pull changes including OSCAL changes

`git pull --recurse-submodules`

#### Push changes including OSCAL changes

`git push --recurse-submodules=check` _then_ `git push --recurse-submodules=on-demand`

#### Helpful config settings:

* `git config diff.submodule log`
* `git config status.submodulesummary 1`
* `git config push.recurseSubmodules check`
* `git config submodule.recurse true`
