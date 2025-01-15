package drive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type EnvCfg struct {
	*JobResData
	*VcapAppData
	*ServicesData

	JobResFile   string `env:"JOB_RESPONSE_FILE"`
	VcapAppJSON  string `env:"VCAP_APPLICATION"`
	ServicesJSON string `env:"CUSTOM_ENV_CI_JOB_SERVICES"`

	ContainerId string `env:"CONTAINER_ID"`

	RegUser string `env:"CUSTOM_ENV_CI_REGISTRY_USER"`
	RegPass string `env:"CUSTOM_ENV_CI_REGISTRY_PASSWORD"`
	JobImg  string `env:"CUSTOM_ENV_CI_JOB_IMAGE"`

	DockerToken string `env:"DOCKER_HUB_TOKEN"`
	DockerUser  string `env:"DOCKER_HUB_USER"`
	DockerPass  string `env:"CF_DOCKER_PASSWORD"`

	WorkerMem  string `env:"WORKER_MEMORY"`
	WorkerDisk string `env:"WORKER_DISK_SIZE"`
}

type JobResData struct {
	Image     JobResImg
	Variables JobResVars
	Services  []JobResSvcs
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
type JobResSvcs struct {
	JobResImg
	Variables JobResVars
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

func (c *EnvCfg) parseJobResFile() (err error) {
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

func (c *EnvCfg) parseVcapAppJSON() (err error) {
	c.VcapAppData, err = parseCfgJSON([]byte(c.VcapAppJSON), &VcapAppData{})
	return err
}

func (c *EnvCfg) parseServicesJSON() (err error) {
	c.ServicesData, err = parseCfgJSON([]byte(c.ServicesJSON), &ServicesData{})
	return err
}

// This is a pretty simple implementation, if our needs get more
// complex we should use one of several existing packages to do this.
// e.g., https://pkg.go.dev/github.com/caarlos0/env/v11
func (c *EnvCfg) parseEnv() *EnvCfg {
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

func getEnvCfg() *EnvCfg {
	cfg := (&EnvCfg{}).parseEnv()
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
