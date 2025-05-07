# Worker Setup

The worker setup requires a number of standalone binaries, some packaged in
tarballs. These compiled assets are not currently version controlled or checked
into LFS because we have not yet established any kind of real pipeline for them.
Probably I will check them into LFS before this gets merged to main.

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
