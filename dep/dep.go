package dep

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
)

const Dependency = "dep"

type Contributor struct {
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	_, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	return Contributor{}, true, nil
}

func (c Contributor) Contribute() error {
	return nil
}
