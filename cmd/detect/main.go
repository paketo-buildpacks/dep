package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"gopkg.in/yaml.v2"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/pkg/errors"
)

const ErrorMsg = "no Gopkg.toml found at root level"
const EmptyTargetEnvVariableMsg = "BP_GO_TARGETS set but with empty value"

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
		return detect.FailStatusCode, fmt.Errorf(ErrorMsg)
	}

	bpYmlFile := filepath.Join(context.Application.Root, "buildpack.yml")
	if exists, err := helper.FileExists(bpYmlFile); err != nil {
		return detect.FailStatusCode, errors.Wrap(err, fmt.Sprintf("error checking filepath: %s", bpYmlFile))
	} else if exists {
		config, err := parseBuildpackYml(bpYmlFile)
		if err != nil {
			return detect.FailStatusCode, errors.Wrap(err, "error reading buildpack.yml")
		}
		if config.Go.ImportPath == "" {
			return context.Fail(), nil
		}

		if environmentTargets, ok := os.LookupEnv("BP_GO_TARGETS"); ok {
			if environmentTargets == "" {
				return detect.FailStatusCode, errors.New(EmptyTargetEnvVariableMsg)
			}
			var targets []string
			for _, target := range strings.Split(environmentTargets, ":") {
				targets = append(targets, target)
			}
			config.Go.Targets = targets
		}

		return context.Pass(buildplan.BuildPlan{
			dep.Dependency: buildplan.Dependency{
				Metadata: buildplan.Metadata{
					"build":       true,
					"import-path": config.Go.ImportPath,
					"targets": config.Go.Targets,
				},
			},
		})
	}
	return context.Fail(), nil
}

type BpYML struct {
	Go         struct {
		ImportPath string `yaml:"import-path"`
		Targets []string `yaml:"targets"`
	} `yaml:"go"`
}

func parseBuildpackYml(bpYmlFilePath string) (BpYML, error) {

	bpYML := BpYML{}
	contents, err := ioutil.ReadFile(bpYmlFilePath)
	if err != nil {
		return bpYML, err
	}

	if err := yaml.Unmarshal(contents, &bpYML); err != nil {
		return bpYML, err
	}
	return bpYML, nil
}
