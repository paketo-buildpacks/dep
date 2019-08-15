package main

import (
	"fmt"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"os"

	"github.com/cloudfoundry/dep-cnb/dep"
	"github.com/cloudfoundry/dep-cnb/utils"

	"github.com/cloudfoundry/libcfbuildpack/build"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default build context: %s", err)
		os.Exit(101)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.BodyError("failure running build: %s", err.Error())
	}

	os.Exit(code)

}

func runBuild(context build.Build) (int, error) {
	context.Logger.Title(context.Buildpack)

	runner := &utils.Command{}

	depContributor, willContribute, err := dep.NewContributor(context, runner)
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := depContributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildpackplan.Plan{})
}
