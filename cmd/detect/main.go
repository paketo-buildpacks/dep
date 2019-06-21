package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/pkg/errors"
)

const MissingGopkgErrorMsg = "no Gopkg.toml found at root level"
const MissingBuildpackYmlErrorMsg = "no buildpack.yml found at root level"
const MissingImportPathErrorMsg = "no import-path found in buildpack.yml"
const EmptyTargetEnvVariableMsg = "BP_GO_TARGETS set but with empty value"


type BuildpackYAML struct {
	Config Config `yaml:"go"`
}

type Config struct {
	ImportPath string `yaml:"import-path"`
	Targets []string `yaml:"targets"`
}

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed detection: %s", err)
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	goPkgFile := filepath.Join(context.Application.Root, "Gopkg.toml")

	if exists, err := helper.FileExists(goPkgFile); err != nil {
		return detect.FailStatusCode, errors.Wrap(err, fmt.Sprintf("error checking filepath: %s", goPkgFile))
	} else if !exists {
		return detect.FailStatusCode, fmt.Errorf(MissingGopkgErrorMsg)
	}

	bpYmlFilePath := filepath.Join(context.Application.Root, "buildpack.yml")
	if exists, err := helper.FileExists(bpYmlFilePath); err != nil {
		return detect.FailStatusCode, errors.Wrap(err, fmt.Sprintf("error checking filepath: %s", bpYmlFilePath))
	} else if exists {
		buildpackYaml := BuildpackYAML{}
		if err := helper.ReadBuildpackYaml(bpYmlFilePath, &buildpackYaml); err != nil {
			return detect.FailStatusCode, errors.Wrap(err, "error reading buildpack.yml")
		}

		if buildpackYaml.Config.ImportPath == "" {
			return context.Fail(), errors.New(MissingImportPathErrorMsg)
		}

		if environmentTargets, ok := os.LookupEnv("BP_GO_TARGETS"); ok {
			if environmentTargets == "" {
				return detect.FailStatusCode, errors.New(EmptyTargetEnvVariableMsg)
			}
			var targets []string
			for _, target := range strings.Split(environmentTargets, ":") {
				targets = append(targets, target)
			}
			buildpackYaml.Config.Targets = targets
		}

		return context.Pass(buildplan.BuildPlan{
			dep.Dependency: buildplan.Dependency{
				Metadata: buildplan.Metadata{
					"build":       true,
					"import-path": buildpackYaml.Config.ImportPath,
					"targets": buildpackYaml.Config.Targets,
				},
			},
		})
	}
	return context.Fail(), errors.New(MissingBuildpackYmlErrorMsg)
}
