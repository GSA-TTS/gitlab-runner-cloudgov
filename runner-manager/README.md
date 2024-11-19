The contents of this directory will be placed in `/home/vcap/app`.

Files:
* **.profile** - Executed at the end of application instance (runtime) start up
  by CloudFoundry. All commands needed to register the runner are in this file
  including the `gitlab-runner register` command and arguments.
* **apt.yml** - Build time definition for installing prerequisite packages for
  the runner manager.
* **cf-runner-config.toml** - Additional configuration for `cf` executor.
  Static runner configuration should be placed here.
