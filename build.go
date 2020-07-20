package dep

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
type BuildPlanRefinery interface {
	BillOfMaterials(postal.Dependency) packit.BuildpackPlanEntry
}

func Build(
	entries EntryResolver,
	dependencies DependencyManager,
	planRefinery BuildPlanRefinery,
	logger LogEmitter,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		entry := entries.Resolve(context.Plan.Entries)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, entry.Version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// todo log SelectedDependency
		bom := planRefinery.BillOfMaterials(dependency)

		depLayer, err := context.Layers.Get(Dep)
		if err != nil {
			return packit.BuildResult{}, err
		}

		depLayer.Launch = entry.Metadata["launch"] == true
		depLayer.Build = entry.Metadata["build"] == true
		depLayer.Cache = entry.Metadata["build"] == true

		// todo check for possible layer reuse

		logger.Process("Executing build process")

		err = depLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess("Installing Dep")
		err = dependencies.Install(dependency, context.CNBPath, depLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{bom},
			},
			Layers: []packit.Layer{
				depLayer,
			},
		}, nil
	}
}
