package drive

import (
	"fmt"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
)

type stage struct {
	// conf
	prep *prepStage
	// run
	// clean

	common commonStage
}

type commonStage struct {
	client *cloudgov.Client
	config *JobConfig
}

func newStage() (s *stage, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error creating stage: %w", err)
		}
	}()

	s.common.client, err = cloudgov.New(&cloudgov.CFClientAPI{}, nil)
	if err != nil {
		return
	}

	s.common.config, err = getJobConfig()
	if err != nil {
		return
	}

	// conf
	s.prep = (*prepStage)(&s.common)
	// run
	// clean

	return
}
