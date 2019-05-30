package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
		importPath, err := parseImportPath(bpYmlFile)
		if err != nil {
			return detect.FailStatusCode, errors.Wrap(err, "error reading buildpack.yml")
		}
		if importPath == "" {
			return context.Fail(), nil
		}
		return context.Pass(buildplan.BuildPlan{
			dep.Dependency: buildplan.Dependency{
				Metadata: buildplan.Metadata{
					"build":       true,
					"import-path": importPath,
				},
			},
		})
	}
	return context.Fail(), nil
}

func parseImportPath(bpYmlFilePath string) (string, error) {
	contents, err := ioutil.ReadFile(bpYmlFilePath)
	if err != nil {
		return "", err
	}
	bpYML := struct {
		ImportPath string `yaml:"import-path"`
	}{}
	if err := yaml.Unmarshal(contents, &bpYML); err != nil {
		return "", err
	}
	return bpYML.ImportPath, nil
}
