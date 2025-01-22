package drive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type JobConfig struct {
	*JobResponse           // Parsed JSON from JobResponseFile
	JobResponseFile string `env:"JOB_RESPONSE_FILE"`

	*VcapAppData
	VcapAppJSON string `env:"VCAP_APPLICATION"`

	// We combine the following to make the container ID.
	// Some are available in JOB_RESPONSE_FILE, but several are only found
	// within `.variables` so we pull all from ENV for consistency.
	ContainerID         string
	JobID               string `env:"CUSTOM_ENV_CI_JOB_ID"`
	RunnerID            string `env:"CUSTOM_ENV_CI_RUNNER_ID"`
	ProjectID           string `env:"CUSTOM_ENV_CI_PROJECT_ID"`
	ConcurrentProjectID string `env:"CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID"`

	CIRegistryUser string `env:"CUSTOM_ENV_CI_REGISTRY_USER"`
	CIRegistryPass string `env:"CUSTOM_ENV_CI_REGISTRY_PASSWORD"`

	DockerHubUser  string `env:"DOCKER_HUB_USER"`
	DockerHubToken string `env:"DOCKER_HUB_TOKEN"`

	WorkerMemory   string `env:"WORKER_MEMORY"`
	WorkerDiskSize string `env:"WORKER_DISK_SIZE"`
}

type JobResponse struct {
	Image     *Image
	Variables []*CIVar
	Services  []*Service
}
type Image struct {
	Name       string
	Alias      string
	Command    []string
	Entrypoint []string
}
type CIVar struct {
	Key   string
	Value string
}
type Service struct {
	*Image
	Variables []*CIVar
}

type VcapAppData struct {
	OrgName   string `json:"organization_name"`
	SpaceName string `json:"space_name"`
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

func (c *JobConfig) parseJobResponseFile() (err error) {
	if c.JobResponseFile == "" {
		return nil
	}

	j, err := os.ReadFile(c.JobResponseFile)
	if err != nil {
		return fmt.Errorf("error reading JobResponseFile: %w", err)
	}

	c.JobResponse, err = parseCfgJSON(j, &JobResponse{})
	return err
}

func (c *JobConfig) parseVcapAppJSON() (err error) {
	c.VcapAppData, err = parseCfgJSON([]byte(c.VcapAppJSON), &VcapAppData{})
	return err
}

// This is a pretty simple implementation, if our needs get more
// complex we should use one of several existing packages to do this.
// e.g., https://pkg.go.dev/github.com/caarlos0/env/v11
func (c *JobConfig) parseEnv() *JobConfig {
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

func GetJobConfig() *JobConfig {
	cfg := (&JobConfig{}).parseEnv()

	if err := cfg.parseJobResponseFile(); err != nil {
		panic(err)
	}
	if err := cfg.parseVcapAppJSON(); err != nil {
		panic(err)
	}
	return cfg
}
