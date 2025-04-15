package drive

import (
	"fmt"
	"os/exec"
	"strings"

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
	*stage
	client *cloudgov.Client
	config *JobConfig
}

func newStage(client *cloudgov.Client) (s *stage, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error creating stage: %w", err)
		}
	}()

	s = &stage{}
	s.common.stage = s

	if client != nil {
		s.common.client = client
	} else {
		s.common.client, err = cloudgov.New(&cloudgov.CFClientAPI{}, nil)
		if err != nil {
			return
		}
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

func (s *stage) RunSSH(guid string, cmd string) error {
	cfg := s.common.config.EgressProxyConfig

	pass, err := s.common.client.SSHCode()
	if err != nil {
		return err
	}

	sshCmd := exec.Command(
		"sshpass", "-p", pass,
		"ssh -p 2222 -T",
		"-o 'StrictHostKeyChecking=no'",
		fmt.Sprintf("-o 'ProxyCommand corkscrew %v %v %%h %%p %v'",
			cfg.ProxyHostSSH, cfg.ProxyPortSSH, cfg.ProxyAuthFile,
		),
		fmt.Sprintf("cf:%s/0@ssh.fr.cloud.gov", guid),
		"cmd",
	)

	fmt.Print(sshCmd.String())
	fmt.Print(strings.Join(sshCmd.Environ(), "\n"))

	out, err := sshCmd.Output()
	if err != nil {
		return err
	}

	fmt.Print(string(out))

	return nil
}
