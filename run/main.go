package main

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/dep"
)

func main() {
	packit.Run(dep.Detect(), dep.Build())
}
