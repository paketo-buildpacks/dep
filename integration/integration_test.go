package integration_test

import (
	"github.com/cloudfoundry/dep-cnb/dep"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	depURI, goURI string
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	bpDir, err := dagger.FindBPRoot()
	Expect(err).NotTo(HaveOccurred())

	depURI, err = dagger.PackageBuildpack(bpDir)
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(depURI)

	goURI, err = dagger.GetLatestBuildpack("go-cnb")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(goURI)

	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	var Expect func(interface{}, ...interface{}) GomegaAssertion

	it.Before(func() {
		Expect = NewWithT(t).Expect
	})

	it("should successfully build a simple app", func() {
		appRoot := filepath.Join("testdata", "simple_app")

		app, err := dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())
		defer app.Destroy()

		Expect(app.Start()).To(Succeed())
		body, _, err := app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring("Hello, World!"))

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

		app, err := dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())
		defer app.Destroy()

		Expect(app.Start()).To(Succeed())
		body, _, err := app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring("Hello, World!"))

		Expect(app.BuildLogs()).To(MatchRegexp("Dep.*: Contributing to layer"))
	})

	it("uses Gopkg.lock as a lockfile for re-builds", func() {
		appDir := filepath.Join("testdata", "with_lockfile")
		app, err := dagger.PackBuild(appDir, goURI, depURI)
		Expect(err).ToNot(HaveOccurred())
		defer app.Destroy()

		depPrefix := "Dep \\d*\\.\\d*\\.\\d*: "
		depPackagesPrefix := "Dep Packages \\w*: "
		contributeMsg := "Contributing to layer"
		Expect(app.BuildLogs()).To(MatchRegexp(depPrefix + contributeMsg))
		Expect(app.BuildLogs()).To(MatchRegexp(depPackagesPrefix + contributeMsg))

		_, imageID, _, err := app.Info()
		Expect(err).NotTo(HaveOccurred())

		app, err = dagger.PackBuildNamedImage(imageID, appDir, goURI, depURI)
		Expect(err).ToNot(HaveOccurred())

		rebuildLogs := app.BuildLogs()
		reuseMsg := "Reusing cached layer"
		Expect(rebuildLogs).To(MatchRegexp(depPrefix + reuseMsg))
		Expect(rebuildLogs).To(MatchRegexp(depPackagesPrefix + reuseMsg))
		Expect(app.Start()).To(Succeed())

		_, _, err = app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
	})

	it("uses the vendored packages when the app is vendored", func() {
		appDir := filepath.Join("testdata", "vendored_app")
		app, err := dagger.PackBuild(appDir, goURI, depURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.BuildLogs()).To(ContainSubstring("Note: skipping `dep ensure` due to non-empty vendor directory."))

		Expect(app.Start()).To(Succeed())
		_, _, err = app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
	})
}
