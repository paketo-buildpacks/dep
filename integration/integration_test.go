package integration_test

import (
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
	defer os.RemoveAll(depURI)

	goURI, err = dagger.GetLatestBuildpack("go-cnb")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(goURI)

	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, _ spec.G, it spec.S) {
	var Expect func(interface{}, ...interface{}) GomegaAssertion
	it.Before(func() {
		Expect = NewWithT(t).Expect
	})

	it("builds successfully", func() {
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
}
