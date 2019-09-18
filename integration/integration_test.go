package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dep-cnb/dep"

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
	defer dagger.DeleteBuildpack(depURI)

	goURI, err = dagger.GetLatestBuildpack("go-compiler-cnb")
	Expect(err).NotTo(HaveOccurred())
	defer dagger.DeleteBuildpack(goURI)

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

	it("uses the vendored packages when the app is vendored", func() {
		appDir := filepath.Join("testdata", "vendored_app")
		app, err := dagger.PackBuild(appDir, goURI, depURI)
		Expect(err).ToNot(HaveOccurred())

		Expect(app.BuildLogs()).To(ContainSubstring("Note: skipping `dep ensure` due to non-empty vendor directory."))

		Expect(app.Start()).To(Succeed())
		_, _, err = app.HTTPGet("/")
		Expect(err).ToNot(HaveOccurred())
	})

	it("uses updated source code on a rebuild", func() {
		appRoot := filepath.Join("testdata", "with_lockfile")

		app, err := dagger.PackBuild(appRoot, goURI, depURI)
		Expect(err).NotTo(HaveOccurred())
		defer app.Destroy()

		Expect(err).ToNot(HaveOccurred())
		Expect(app.BuildLogs()).To(MatchRegexp("Dep.*: Contributing to layer"))

		_, imageID, _, err := app.Info()
		appRoot = filepath.Join("testdata", "with_lockfile_modified")
		app, err = dagger.PackBuildNamedImage(imageID, appRoot, goURI, depURI)
		Expect(app.Start()).To(Succeed())
		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("The source changed!"))
	})

	when("the app specifies ldflags", func() {
		it.Focus("should build the app with those build flags", func() {
			app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app_with_target_and_ldflags"), goURI, depURI)
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			Expect(app.Start()).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("main.version: v1.2.3"))
			Expect(body).To(ContainSubstring("main.sha: 7a82056"))
		})
	})
}
