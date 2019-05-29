package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("there is no Gopkg.toml", func() {
		it("should fail", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(ErrorMsg))
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("Gopkg.toml exists", func() {
		it("should pass and add to dep to the build plan", func() {
			goPkgString := fmt.Sprintf("This is a go pkg toml")
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gopkg.toml"), goPkgString)

			code, err := runDetect(factory.Detect)
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(err).ToNot(HaveOccurred())

			plan := buildplan.BuildPlan{
				dep.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{"build": true},
				},
			}

			Expect(factory.Output).To(Equal(plan))
		})
	})
}
