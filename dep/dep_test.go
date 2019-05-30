package dep_test

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/dep-cnb/dep"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=dep.go -destination=mocks_test.go -package=dep_test

func TestUnitGoMod(t *testing.T) {
	spec.Run(t, "Go Dep", testDep, spec.Report(report.Terminal{}))
}

func testDep(t *testing.T, when spec.G, it spec.S) {
	var (
		factory     *test.BuildFactory
		mockRunner  *MockRunner
		mockCtrl    *gomock.Controller
		packageName string
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		mockCtrl = gomock.NewController(t)
		mockRunner = NewMockRunner(mockCtrl)
		packageName = "app"
	})

	when("NewContributor", func() {
		it("returns true if it exists in the buildplan", func() {
			factory.AddBuildPlan(dep.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{
					dep.ImportPath: packageName,
				}})

			_, willContribute, err := dep.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			_, willContribute, err := dep.NewContributor(factory.Build, mockRunner)

			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})
	})

	when("ContributeDep", func() {
		it("installs dep when included in the build plan", func() {
			factory.AddBuildPlan(dep.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{
					dep.ImportPath: packageName,
				}})

			stubFixture := filepath.Join("testdata", "stub.tar.gz")
			factory.AddDependency(dep.Dependency, stubFixture)

			contributor, willContribute, err := dep.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
			Expect(contributor.ContributeDep()).To(Succeed())

			layer := factory.Build.Layers.Layer(dep.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
		})
	})

	when("ContributePackages", func() {
		it.Pend("runs dep ensure", func() {
			factory.AddBuildPlan(dep.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{
					dep.ImportPath: packageName,
				}})
			layer := factory.Build.Layers.Layer(dep.Packages)

			installDir := filepath.Join(layer.Root, "src", packageName)
			mockRunner.EXPECT().CustomRun(installDir,
				[]string{fmt.Sprintf("GOPATH=%s", factory.Build.Layers.Layer(dep.Packages).Root)},
				os.Stdout, os.Stderr,
				gomock.Any(), "ensure").Return(nil)

			contributor, willContribute, err := dep.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
			Expect(contributor.ContributePackages()).To(Succeed())
		})
	})

	when("ContributeBinary", func() {
		it("runs go install", func() {
			factory.AddBuildPlan(dep.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{
					dep.ImportPath: packageName,
				}})
			appBinaryLayer := factory.Build.Layers.Layer(dep.AppBinary)
			appBinaryLayer.Touch()
			packagesLayer := factory.Build.Layers.Layer(dep.Packages)
			installDir := filepath.Join(packagesLayer.Root, "src", packageName)

			mockRunner.EXPECT().CustomRun(installDir, []string{
				fmt.Sprintf("GOPATH=%s", packagesLayer.Root),
				fmt.Sprintf("GOBIN=%s", appBinaryLayer.Root),
			}, os.Stdout, os.Stderr, "go", "install", "-buildmode", "pie", "-tags", "cloudfoundry")
			contributor, willContribute, err := dep.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
			Expect(contributor.ContributeBinary()).To(Succeed())
		})
	})

	when("deleteAppDir", func() {
		it("succesfully deletes all contents of appDir", func() {
			dummyFile := filepath.Join(factory.Build.Application.Root, "dummy")
			Expect(ioutil.WriteFile(dummyFile, []byte("baller"), 0777))
			factory.AddBuildPlan(dep.Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{
					dep.ImportPath: packageName,
				}})
			contributor, willContribute, err := dep.NewContributor(factory.Build, mockRunner)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())

			Expect(contributor.DeleteAppDir()).To(Succeed())
			appDirContents, err := ioutil.ReadDir(factory.Build.Application.Root)
			Expect(err).NotTo(HaveOccurred())
			Expect(appDirContents).To(BeEmpty())

		})
	})
}
