package dep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	Dependency = "dep"
	Packages   = "packages"
	lockFile   = "Gopkg.lock"
	AppBinary  = "app-binary"
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
	_, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	contributor := Contributor{
		context:        context,
		runner:         runner,
		packagesLayer:  context.Layers.Layer(Packages),
		appBinaryLayer: context.Layers.Layer(AppBinary),
	}

	contributor.appDirName = "app"
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

	return c.ContributeStartCommand()
}

func (c *Contributor) ContributeDep() error {
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
		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		return os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), c.depLayer.Root))
	}, layers.Build, layers.Cache)
}

func (c *Contributor) ContributePackages() error {
	if err := c.setPackagesMetadata(); err != nil {
		return err
	}

	return c.packagesLayer.Contribute(c.goDepPackages, func(layer layers.Layer) error {
	if err := os.MkdirAll(c.installDir, 0777); err != nil {
			return err
		}

		if err := helper.CopyDirectory(c.context.Application.Root, c.installDir); err != nil {
			return err
		}

		return c.runner.CustomRun(c.installDir, []string{fmt.Sprintf("GOPATH=%s", c.packagesLayer.Root)}, os.Stdout, os.Stderr, "dep", "ensure")
	}, layers.Cache)
}

func (c *Contributor) ContributeBinary() error {
	return c.appBinaryLayer.Contribute(getAppBinaryMetadata(), func(layer layers.Layer) error {
		layer.Logger.SubsequentLine("Running `go install`")
		args := []string{"install", "-buildmode", "pie", "-tags", "cloudfoundry"}

		return c.runner.CustomRun(c.installDir, []string{
			fmt.Sprintf("GOPATH=%s", c.packagesLayer.Root),
			fmt.Sprintf("GOBIN=%s", layer.Root),
		}, os.Stdout, os.Stderr, "go", args...)
	}, layers.Launch)
}

func (c *Contributor) ContributeStartCommand() error {
	appBinaryPath := filepath.Join(c.appBinaryLayer.Root, c.appDirName)
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

func getAppBinaryMetadata() Identifiable {
	timeNow := strconv.FormatInt(time.Now().UnixNano(), 32)
	hash := sha256.Sum256([]byte(timeNow))
	checksum := hex.EncodeToString(hash[:])
	return Identifiable{Name: "Binary", Checksum: checksum}
}
