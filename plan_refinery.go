package dep

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

type PlanRefinery struct{}

func NewPlanRefinery() PlanRefinery {
	return PlanRefinery{}
}

func (r PlanRefinery) BillOfMaterials(dependency postal.Dependency) packit.BuildpackPlanEntry {
	return packit.BuildpackPlanEntry{
		Name: dependency.ID,
		Metadata: map[string]interface{}{
			"version":  dependency.Version,
			"licenses": []string{},
			"name":     dependency.Name,
			"sha256":   dependency.SHA256,
			"stacks":   dependency.Stacks,
			"uri":      dependency.URI,
		},
	}
}
