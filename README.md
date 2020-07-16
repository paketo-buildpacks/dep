# Dep Cloud Native Buildpack

The Dep CNB provides the
[`dep`](https://golang.github.io/dep/docs/introduction.html) executable. The
buildpack installs dep onto the `$PATH` which makes it available for subsequent
buildpacks.

## Integration

The Dep CNB provides `dep` as a dependency. Downstream
buildpacks can require the dep dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Dep dependency is "dep". This value is
  # considered part of the public API for the buildpack and will not change
  # without a plan for deprecation.
  name = "dep"

  # Note: The version field is unsupported as there is no version for a set of
  # dep.

  # The Dep buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Dep
    # dependency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Dep
    # during its build process, this flag should be set to true.
    build = true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configuration

The dep buildpack does not support configurations via `buildpack.yml`.
