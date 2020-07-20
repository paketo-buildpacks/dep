package dep_test

import (
	"bytes"
	"testing"

	dep "github.com/paketo-buildpacks/dep"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanEntryResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer   *bytes.Buffer
		resolver dep.PlanEntryResolver
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		resolver = dep.NewPlanEntryResolver(dep.NewLogEmitter(buffer))
	})

	context("when entry flags differ", func() {
		context("OR's them together on best plan entry", func() {
			it("has all flags", func() {
				entry := resolver.Resolve([]packit.BuildpackPlanEntry{
					{
						Name: "dep",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "dep",
						Metadata: map[string]interface{}{
							"build": true,
						},
					},
				})
				Expect(entry).To(Equal(packit.BuildpackPlanEntry{
					Name: "dep",
					Metadata: map[string]interface{}{
						"build":  true,
						"launch": true,
					},
				}))
			})
		})
	})
}
