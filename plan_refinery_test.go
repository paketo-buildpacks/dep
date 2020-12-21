package dep_test

import (
	"testing"

	dep "github.com/paketo-buildpacks/dep"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanRefinery(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		refinery dep.PlanRefinery
	)

	it.Before(func() {
		refinery = dep.NewPlanRefinery()
	})

	context("BillOfMaterials", func() {
		it("returns a refined build plan entry", func() {
			entry := refinery.BillOfMaterials(postal.Dependency{
				ID:      "some-id",
				Name:    "some-name",
				Stacks:  []string{"some-stack"},
				URI:     "some-uri",
				SHA256:  "some-sha",
				Version: "some-version",
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name: "some-id",
				Metadata: map[string]interface{}{
					"version":  "some-version",
					"licenses": []string{},
					"name":     "some-name",
					"sha256":   "some-sha",
					"stacks":   []string{"some-stack"},
					"uri":      "some-uri",
				},
			}))
		})
	})
}
