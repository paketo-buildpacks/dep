package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"

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

	when("Gopkg.toml exists and buildpack.yml does not exist", func() {
		it("should fail detection", func() {
			goPkgString := fmt.Sprintf("This is a go pkg toml")
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gopkg.toml"), goPkgString)

			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("Gopkg.toml exists and buildpack.yml specifies an `import-path` and go targets", func() {

		var bpYmlString string

		it.Before(func() {
			bpYmlString = `---

go:
  import-path: some/app
  targets: ["./path/to/first", "./path/to/second"]`
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), bpYmlString)
			goPkgString := fmt.Sprintf("This is a go pkg toml")
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gopkg.toml"), goPkgString)
		})

		it("adds the `import-path` and targets to the build plan", func() {

			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))

			plan := buildplan.BuildPlan{
				dep.Dependency: buildplan.Dependency{
					Metadata: buildplan.Metadata{
						"build": true,
						"import-path": "some/app",
						"targets": []string{"./path/to/first", "./path/to/second"},
					},
				},
			}

			Expect(factory.Output).To(Equal(plan))
		})

		when("BP_GO_TARGETS environment variable is set", func() {
			it("should use the BP_GO_TARGETS value in the build plan", func() {

				err := os.Setenv("BP_GO_TARGETS", "./path/to/third:./path/to/fourth")
				Expect(err).NotTo(HaveOccurred())

				code, err := runDetect(factory.Detect)
				Expect(err).ToNot(HaveOccurred())
				Expect(code).To(Equal(detect.PassStatusCode))
				plan := buildplan.BuildPlan{
					dep.Dependency: buildplan.Dependency{
						Metadata: buildplan.Metadata{
							"build": true,
							"import-path": "some/app",
							"targets": []string{"./path/to/third", "./path/to/fourth"},
						},
					},
				}
				Expect(factory.Output).To(Equal(plan))

			})
		})

		when("BP_GO_TARGETS environment variable is set but empty", func() {
			it("should use fail the detect phase", func() {

				err := os.Setenv("BP_GO_TARGETS", "")
				Expect(err).NotTo(HaveOccurred())

				code, err := runDetect(factory.Detect)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(EmptyTargetEnvVariableMsg))
				Expect(code).To(Equal(detect.FailStatusCode))

			})
		})

	})

	when("Gopkg.toml exists and buildpack.yml empty", func() {
		it("fails to detect", func() {
			bpYmlString := ""
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), bpYmlString)

			goPkgString := fmt.Sprintf("This is a go pkg toml")
			test.WriteFile(t, filepath.Join(factory.Detect.Application.Root, "Gopkg.toml"), goPkgString)

			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})
}
