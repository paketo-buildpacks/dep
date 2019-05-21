package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/pkg/errors"
)

const ErrorMsg = "no Gopkg.toml found at root level"

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	goPkgFile := filepath.Join(context.Application.Root, "Gopkg.toml")

	if exists, err := helper.FileExists(goPkgFile); err != nil {
		return detect.FailStatusCode, errors.Wrap(err, fmt.Sprintf("error checking filepath: %s", goPkgFile))
	} else if !exists {
		return detect.FailStatusCode, fmt.Errorf(ErrorMsg)
	}

	return context.Pass(buildplan.BuildPlan{
		dep.Dependency: buildplan.Dependency{
			Metadata: buildplan.Metadata{
				"build": true,
			},
		},
	})
}
