package dep

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve(name string, entries []packit.BuildpackPlanEntry, priorites []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch bool, build bool)
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
	entryResolver EntryResolver,
	dependencies DependencyManager,
	planRefinery BuildPlanRefinery,
	clock chronos.Clock,
	logger scribe.Emitter,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		entry, _ := entryResolver.Resolve(Dep, context.Plan.Entries, nil)

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = "default"
		}

		dependency, err := dependencies.Resolve(
			filepath.Join(context.CNBPath, "buildpack.toml"),
			entry.Name,
			version,
			context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := planRefinery.BillOfMaterials(dependency)

		depLayer, err := context.Layers.Get(Dep)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := depLayer.Metadata[DependencyCacheKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", depLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{bom},
				},
				Layers: []packit.Layer{depLayer},
			}, nil
		}

		logger.Process("Executing build process")

		depLayer, err = depLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		depLayer.Launch, depLayer.Build = entryResolver.MergeLayerTypes(Dep, context.Plan.Entries)
		depLayer.Cache = depLayer.Build

		logger.Subprocess("Installing Dep")

		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, depLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		depLayer.Metadata = map[string]interface{}{
			DependencyCacheKey: dependency.SHA256,
			"built_at":         clock.Now().Format(time.RFC3339Nano),
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
