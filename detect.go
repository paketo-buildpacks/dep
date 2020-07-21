package dep

import "github.com/paketo-buildpacks/packit"

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "dep"},
				},
				Requires: nil,
			},
		}, nil
	}
}
