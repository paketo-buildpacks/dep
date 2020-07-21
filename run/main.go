package main

import (
	"os"

	"github.com/paketo-buildpacks/dep"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	logEmitter := dep.NewLogEmitter(os.Stdout)
	entryResolver := dep.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := dep.NewPlanRefinery()

	packit.Run(
		dep.Detect(),
		dep.Build(
			entryResolver,
			dependencyManager,
			planRefinery,
			chronos.DefaultClock,
			logEmitter,
		),
	)
}
