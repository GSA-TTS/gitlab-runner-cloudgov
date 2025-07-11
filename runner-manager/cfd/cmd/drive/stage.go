package drive

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner-manager/cfd/cloudgov"
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

	s.common.config, err = getJobConfig()
	if err != nil {
		return
	}

	if client != nil {
		s.common.client = client
	} else {
		s.common.client, err = cloudgov.New(
			&cloudgov.CFClientAPI{},
			&cloudgov.Opts{APIRootURL: s.common.config.CFApi},
		)
		if err != nil {
			return
		}
	}

	// conf
	s.prep = (*prepStage)(&s.common)
	// run
	// clean

	return
}

func (s *stage) RunSSH(guid string, cmd string) error {
	pass, err := s.common.client.SSHCode()
	if err != nil {
		return err
	}

	args := []string{"ssh", "-p 2222", "-T", "-o StrictHostKeyChecking=no"}
	host := fmt.Sprintf("cf:%s/0@ssh.fr-stage.cloud.gov", guid)

	epCfg := s.common.config.EgressProxyConfig
	if epCfg != (EgressProxyConfig{}) {
		proxy := fmt.Sprintf("-o ProxyCommand corkscrew %v %v %%h %%p %v",
			epCfg.ProxyHostSSH, epCfg.ProxyPortSSH, epCfg.ProxyAuthFile,
		)
		args = append(args, proxy)
	}

	sshCmd := exec.Command("sshpass", append(args, host)...)
	sshCmd.Stdin = strings.NewReader(pass) // give pass to sshpass through stdin

	out, err := sshCmd.Output()
	if err != nil {
		return err
	}

	fmt.Print(string(out))

	return nil
}
