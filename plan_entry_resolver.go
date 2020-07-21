package dep

import (
	"github.com/paketo-buildpacks/packit"
)

type PlanEntryResolver struct {
	logger LogEmitter
}

func NewPlanEntryResolver(logger LogEmitter) PlanEntryResolver {
	return PlanEntryResolver{
		logger: logger,
	}
}

func (r PlanEntryResolver) Resolve(entries []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {

	// there's no data right now to heuristically pick an entry
	chosenEntry := entries[0]

	if chosenEntry.Metadata == nil {
		chosenEntry.Metadata = map[string]interface{}{}
	}

	for _, entry := range entries {
		if entry.Metadata["build"] == true {
			chosenEntry.Metadata["build"] = true
		}

		if entry.Metadata["launch"] == true {
			chosenEntry.Metadata["launch"] = true
		}
	}

	return chosenEntry
}
