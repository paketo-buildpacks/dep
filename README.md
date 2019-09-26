# Dep Cloud Native Buildpack

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## Required buildpack.yml

The `dep-cnb` requires a `buildpack.yml` file in the root of the application directory, and must include `go.import-path` for the application:

```yaml
go:
  import-path: hubgit.net/user/app
```

See `integration/testdata/` subfolders for examples.
