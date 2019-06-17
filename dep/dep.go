package dep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

const (
	Dependency = "dep"
	Packages   = "packages"
	lockFile   = "Gopkg.lock"
	AppBinary  = "app-binary"
	ImportPath = "import-path"
	Targets    = "targets"
)

type Contributor struct {
	context        build.Build
	runner         Runner
	depLayer       layers.DependencyLayer
	packagesLayer  layers.Layer
	appBinaryLayer layers.Layer
	logger         logger.Logger
	goDepPackages  Identifiable
	appDirName     string
	installDir     string
	vendored       bool
	Targets        []string
}

type Identifiable struct {
	Name     string
	Checksum string
}

func (l Identifiable) Identity() (string, string) {
	return l.Name, l.Checksum
}

type Runner interface {
	Run(dir, bin string, args ...string) (string, error)
	RunSilent(dir, bin string, args ...string) (string, error)
	CustomRun(dir string, env []string, out, err io.Writer, bin string, args ...string) error
}

func NewContributor(context build.Build, runner Runner) (Contributor, bool, error) {
	dependency, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	importPath, exists := dependency.Metadata[ImportPath]
	if !exists {
		return Contributor{}, false, nil
	}

	vendored, err := isVendored(context)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{
		context:        context,
		runner:         runner,
		packagesLayer:  context.Layers.Layer(Packages),
		appBinaryLayer: context.Layers.Layer(AppBinary),
		vendored:       vendored,
		logger:         context.Logger,
	}

	targets, exists := dependency.Metadata[Targets]
	if exists {
		if targets, ok := targets.([]string); ok {
			contributor.Targets = targets
		}
	}

	appDirName, ok := importPath.(string)
	if !ok {
		return Contributor{}, false, nil
	}

	contributor.appDirName = appDirName
	contributor.installDir = filepath.Join(contributor.packagesLayer.Root, "src", contributor.appDirName)

	return contributor, true, nil
}

func (c *Contributor) Contribute() error {
	if err := c.ContributeDep(); err != nil {
		return err
	}

	if err := c.ContributePackages(); err != nil {
		return err
	}

	if err := c.ContributeBinary(); err != nil {
		return err
	}

	if err := c.ContributeStartCommand(); err != nil {
		return err
	}

	return c.DeleteAppDir()
}

func (c *Contributor) ContributeDep() error {
	if c.vendored {
		c.logger.Info("Note: skipping dep installation due to non-empty vendor directory.")
		return nil
	}

	deps, err := c.context.Buildpack.Dependencies()
	if err != nil {
		return err
	}

	dep, err := deps.Best(Dependency, "*", c.context.Stack)
	if err != nil {
		return err
	}

	c.depLayer = c.context.Layers.DependencyLayer(dep)

	return c.depLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		return helper.ExtractTarGz(artifact, layer.Root, 1)
	}, layers.Build, layers.Cache)
}

func (c *Contributor) ContributePackages() error {
	if err := c.setPackagesMetadata(); err != nil {
		return err
	}

	return c.packagesLayer.Contribute(c.goDepPackages, func(layer layers.Layer) error {
		if err := helper.CopyDirectory(c.context.Application.Root, c.installDir); err != nil {
			return err
		}

		if c.vendored {
			c.logger.Info("Note: skipping `dep ensure` due to non-empty vendor directory.")
			return nil
		}

		layer.Logger.SubsequentLine("Fetching any unsaved dependencies (using `dep ensure`)")
		depBin := filepath.Join(c.depLayer.Root, "dep")
		return c.runner.CustomRun(c.installDir,
			[]string{fmt.Sprintf("GOPATH=%s", c.packagesLayer.Root)},
			os.Stdout, os.Stderr,
			depBin, "ensure")

	}, layers.Cache)
}

func (c *Contributor) ContributeBinary() error {
	return c.appBinaryLayer.Contribute(getAppBinaryMetadata(), func(layer layers.Layer) error {
		layer.Logger.SubsequentLine("Running `go install`")
		args := []string{"install", "-buildmode", "pie", "-tags", "cloudfoundry"}

		if len(c.Targets) > 0 {
			args = append(args, c.Targets...)
		}
		return c.runner.CustomRun(c.installDir, []string{
			fmt.Sprintf("GOPATH=%s", c.packagesLayer.Root),
			fmt.Sprintf("GOBIN=%s", layer.Root),
		}, os.Stdout, os.Stderr,
			"go", args...)
	}, layers.Launch)
}

func (c *Contributor) ContributeStartCommand() error {
	appBinaryPath := filepath.Join(c.appBinaryLayer.Root, filepath.Base(c.appDirName))
	return c.context.Layers.WriteApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", appBinaryPath}}})
}

func (c *Contributor) setPackagesMetadata() error {
	meta := Identifiable{"Dep Packages", strconv.FormatInt(time.Now().UnixNano(), 16)}

	if exists, err := helper.FileExists(filepath.Join(c.context.Application.Root, lockFile)); err != nil {
		return err
	} else if exists {
		out, err := ioutil.ReadFile(filepath.Join(c.context.Application.Root, lockFile))
		if err != nil {
			return err
		}

		hash := sha256.Sum256(out)
		meta.Checksum = hex.EncodeToString(hash[:])
	}

	c.goDepPackages = meta
	return nil
}

func (c Contributor) DeleteAppDir() error {
	files, err := ioutil.ReadDir(c.context.Application.Root)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := os.RemoveAll(filepath.Join(c.context.Application.Root, file.Name())); err != nil {
			return err
		}
	}

	return nil
}

func isVendored(context build.Build) (bool, error) {
	vendorPath := filepath.Join(context.Application.Root, "vendor")
	vendorDirExists, err := helper.FileExists(vendorPath)
	if err != nil {
		return false, err
	}

	if vendorDirExists {
		files, err := ioutil.ReadDir(vendorPath)
		if err != nil {
			return false, err
		}

		for _, file := range files {
			if file.IsDir() {
				return true, nil
			}
		}
	}

	return false, nil
}

func getAppBinaryMetadata() Identifiable {
	timeNow := strconv.FormatInt(time.Now().UnixNano(), 32)
	hash := sha256.Sum256([]byte(timeNow))
	checksum := hex.EncodeToString(hash[:])
	return Identifiable{Name: "App Binary", Checksum: checksum}
}
