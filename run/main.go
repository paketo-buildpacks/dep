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

	packit.Run(
		dep.Detect(),
		dep.Build(
			dep.NewPlanEntryResolver(logEmitter),
			postal.NewService(cargo.NewTransport()),
			dep.NewPlanRefinery(),
			chronos.DefaultClock,
			logEmitter,
		),
	)
}
