# Worker Setup

The worker setup requires a number of standalone binaries, some packaged in
tarballs. Creating the bundle is currently a manual process. Binaries and the git
tarball should be checked into git-lfs.

## Assumed directory structure

```text
├── bundle
│   ├── bin
│   │   ├── git-lfs
│   │   └── gitlab-runner-helper
│   └── git.tgz
└── tar
```

## bundle

This will be tar-balled into `bundle.tgz` and transferred to the workers.

### bin/git-lfs and bin/gitlab-runner-helper

These are both standalone Go binaries downloaded manually from the package maintainers.

### git.tgz

Tarball of a standalone git packaged by DevTools.

## tar

Standalone `tar` packaged by DevTools, used to unpack `bundle.tgz` and then `git.tgz`.
