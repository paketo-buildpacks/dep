package main

import (
	"os"

	"github.com/paketo-buildpacks/dep"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout)

	packit.Run(
		dep.Detect(),
		dep.Build(
			draft.NewPlanner(),
			postal.NewService(cargo.NewTransport()),
			dep.NewPlanRefinery(),
			chronos.DefaultClock,
			logEmitter,
		),
	)
}
