package drive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type EnvConfig struct {
	*JobResData
	*VcapAppData
	*ServicesData

	JobResFile   string `env:"JOB_RESPONSE_FILE"`
	VcapAppJSON  string `env:"VCAP_APPLICATION"`
	ServicesJSON string `env:"CUSTOM_ENV_CI_JOB_SERVICES"`

	ContainerId string `env:"CONTAINER_ID"`

	JobImg         string `env:"CUSTOM_ENV_CI_JOB_IMAGE"`
	CIRegistryUser string `env:"CUSTOM_ENV_CI_REGISTRY_USER"`
	CIRegistryPass string `env:"CUSTOM_ENV_CI_REGISTRY_PASSWORD"`

	DockerHubUser  string `env:"DOCKER_HUB_USER"`
	DockerHubToken string `env:"DOCKER_HUB_TOKEN"`

	WorkerMemory   string `env:"WORKER_MEMORY"`
	WorkerDiskSize string `env:"WORKER_DISK_SIZE"`
}

type JobResData struct {
	Image     *JobResImg
	Variables *JobResVars
	Services  []*JobResServices
}
type JobResImg struct {
	Name       string
	Alias      string
	Command    []string
	Entrypoint []string
}
type JobResVars []struct {
	Key   string
	Value string
}
type JobResServices struct {
	*JobResImg
	Variables *JobResVars
}

type VcapAppData struct {
	OrgName   string `json:"organization_name"`
	SpaceName string `json:"space_name"`
}

type ServicesData []struct {
	Name       string
	Alias      string
	Entrypoint []string
	Command    []string
}

func parseCfgJSON[R any](j []byte, r *R) (*R, error) {
	if len(j) < 1 {
		return nil, nil
	}
	if err := json.Unmarshal(j, r); err != nil {
		return nil, fmt.Errorf("error parsing %t: %w", r, err)
	}
	return r, nil
}

func (c *EnvConfig) parseJobResFile() (err error) {
	if c.JobResFile == "" {
		return nil
	}

	j, err := os.ReadFile(c.JobResFile)
	if err != nil {
		return fmt.Errorf("error reading JobResFile: %w", err)
	}

	c.JobResData, err = parseCfgJSON(j, &JobResData{})
	return err
}

func (c *EnvConfig) parseVcapAppJSON() (err error) {
	c.VcapAppData, err = parseCfgJSON([]byte(c.VcapAppJSON), &VcapAppData{})
	return err
}

func (c *EnvConfig) parseServicesJSON() (err error) {
	c.ServicesData, err = parseCfgJSON([]byte(c.ServicesJSON), &ServicesData{})
	return err
}

// This is a pretty simple implementation, if our needs get more
// complex we should use one of several existing packages to do this.
// e.g., https://pkg.go.dev/github.com/caarlos0/env/v11
func (c *EnvConfig) parseEnv() *EnvConfig {
	ct := reflect.TypeOf(*c)
	ce := reflect.ValueOf(c).Elem()

	for i := 0; i < ct.NumField(); i++ {
		field := ce.Field(i)
		fieldTag := ct.Field(i).Tag.Get("env")
		if fieldTag == "" || field.Kind() != reflect.String {
			continue
		}
		field.SetString(os.Getenv(fieldTag))
	}

	return c
}

func getEnvCfg() *EnvConfig {
	cfg := (&EnvConfig{}).parseEnv()
	if err := cfg.parseJobResFile(); err != nil {
		panic(err)
	}
	if err := cfg.parseVcapAppJSON(); err != nil {
		panic(err)
	}
	if err := cfg.parseServicesJSON(); err != nil {
		panic(err)
	}
	return cfg
}
