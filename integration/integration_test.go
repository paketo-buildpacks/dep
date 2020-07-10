package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/dep/dep"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	depURI, goURI string
)

func Package(root, version string, cached bool) (string, error) {
	var cmd *exec.Cmd

	bpPath := filepath.Join(root, "artifact")
	if cached {
		cmd = exec.Command(".bin/packager", "--archive", "--version", version, fmt.Sprintf("%s-cached", bpPath))
	} else {
		cmd = exec.Command(".bin/packager", "--archive", "--uncached", "--version", version, bpPath)
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("PACKAGE_DIR=%s", bpPath))
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if cached {
		return fmt.Sprintf("%s-cached.tgz", bpPath), err
	}

	return fmt.Sprintf("%s.tgz", bpPath), err
}

func BeforeSuite() {
	bpDir, err := filepath.Abs("./..")
	Expect(err).NotTo(HaveOccurred())

	depURI, err = Package(bpDir, "1.2.3", false)
	Expect(err).ToNot(HaveOccurred())

	goURI, err = dagger.GetLatestCommunityBuildpack("paketo-buildpacks", "go-compiler")
	Expect(err).ToNot(HaveOccurred())
}

func AfterSuite() {
	Expect(dagger.DeleteBuildpack(depURI)).To(Succeed())
	Expect(dagger.DeleteBuildpack(goURI)).To(Succeed())
}

func TestIntegration(t *testing.T) {
	RegisterTestingT(t)
	BeforeSuite()
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}), spec.Parallel())
	AfterSuite()
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var (
		Expect func(interface{}, ...interface{}) GomegaAssertion
		app    *dagger.App
		err    error
	)

	it.Before(func() {
		Expect = NewWithT(t).Expect
	})

	it.After(func() {
		if app != nil {
			Expect(app.Destroy()).To(Succeed())
		}
	})

	it("should successfully build a simple app", func() {
		appRoot := filepath.Join("testdata", "simple_app")

		app, err = dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())

		Expect(app.Start()).To(Succeed())
		body, _, err := app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring("Hello, World!"))
		Expect(body).To(MatchRegexp(`PATH=.*/layers/paketo-buildpacks_dep/app-binary/bin`))

		Expect(app.BuildLogs()).To(MatchRegexp("Dep.*: Contributing to layer"))
	})

	it("should fail to build if the app does not specify import-path", func() {
		appRoot := filepath.Join("testdata", "simple_app_without_import_path")
		_, err := dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(dep.MissingImportPathErrorMsg))
	})

	it("should successfully build a simple app with target", func() {
		appRoot := filepath.Join("testdata", "simple_app_with_target")

		app, err = dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())

		Expect(app.Start()).To(Succeed())
		body, _, err := app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring("Hello, World!"))

		Expect(app.BuildLogs()).To(MatchRegexp("Dep.*: Contributing to layer"))
	})

	it("uses the vendored packages when the app is vendored", func() {
		appDir := filepath.Join("testdata", "vendored_app")
		app, err = dagger.PackBuild(appDir, goURI, depURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.BuildLogs()).To(ContainSubstring("Note: skipping `dep ensure` due to non-empty vendor directory."))

		Expect(app.Start()).To(Succeed())
		_, _, err = app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
	})

	it("uses updated source code on a rebuild", func() {
		appRoot := filepath.Join("testdata", "with_lockfile")

		app, err = dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())

		Expect(err).ToNot(HaveOccurred())
		Expect(app.BuildLogs()).To(MatchRegexp("Dep.*: Contributing to layer"))

		_, imageID, _, err := app.Info()
		Expect(err).NotTo(HaveOccurred())

		appRoot = filepath.Join("testdata", "with_lockfile_modified")
		app, err = dagger.PackBuildNamedImage(imageID, appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())
		Expect(app.Start()).To(Succeed())

		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("The source changed!"))
	})

	when("the app specifies ldflags", func() {
		it("should build the app with those build flags", func() {
			app, err = dagger.PackBuild(filepath.Join("testdata", "simple_app_with_target_and_ldflags"), goURI, depURI)
			Expect(err).ToNot(HaveOccurred())

			Expect(app.Start()).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("main.version: v1.2.3"))
			Expect(body).To(ContainSubstring("main.sha: 7a82056"))
		})
	})
}
