# CloudFoundry (cloud.gov) driver

This is a _driver_ developed for use with `gitlab-runner`’s [Custom executor](https://docs.gitlab.com/runner/executors/custom.html) to prepare, run, and clean up runner managers, workers, and services.

## Go setup

### Install Go

You can use Homebrew to install Go on MacOS (it is generally unnecessary to use a version manager). You could also download Go packages for Linux, Mac, and Windows from [Go's website](https://go.dev/doc/install).

```sh
brew install go
```

### Manage dependencies

Manage project dependencies with `go` & `go mod`. Go programs are divided first into _modules_ and then into _packages_. Dependencies are listed in `./go.mod`.

#### To download dependencies

To download deps as listed in `./go.mod`, run:

```sh
go mod download
```

Or, to walk the project and as needed install & remove modules while updating `./go.mod` & `./go.sum` accordingly:

```sh
go mod tidy
```

#### To `get` new dependencies

While `go mod tidy` can install dependencies you've already imported in your packages, you can also install them explicitly:

```sh
go get github.com/google/some-mod/pkg
```

## Running tests

### Unit tests

The simplest way to run tests—or the one with the least typing, at least—is with `make`.

```sh
make test
```

### Integration tests

We only have one integration test right now, and to get it running you'll need to do a bit of local setup.

1. You will need to first get a username & password for some space on cloud.gov that has at least one app.
1. Then you can add those credentials to `./cg/testdata/.cg_creds` in the style of the `.cg_creds.sample` file there.
1. Run the test with `make integration`, which should give you an error and, in its output, show you what the resulting JSON looks like.
1. Copy that JSON result over to the last line of your `.cg_creds` file and run `make integration` again, this time it should succeed.

## Builds

You can also run a build with `make` (`make build`, or simply `make`), but it won't do you much good because we aren't set up to do anything with an executable yet.
