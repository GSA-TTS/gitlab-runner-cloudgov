package drive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
)

type JobConfig struct {
	*JobResponse           // Parsed JSON from JobResponseFile
	JobResponseFile string `env:"JOB_RESPONSE_FILE"`

	*VcapAppData
	VcapAppJSON string `env:"VCAP_APPLICATION"`

	Manifest *cloudgov.AppManifest

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
type Service struct {
	*Image
	Variables []*CIVar
	Manifest  *cloudgov.AppManifest
	Config    *JobConfig
}
type CIVar struct {
	Key   string
	Value string
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
	c.JobResponse = &JobResponse{}

	if c.JobResponseFile == "" {
		return nil
	}

	j, err := os.ReadFile(c.JobResponseFile)
	if err != nil {
		return fmt.Errorf("error reading JobResponseFile: %w", err)
	}

	c.JobResponse, err = parseCfgJSON(j, c.JobResponse)
	if err != nil {
		return err
	}

	for _, s := range c.Services {
		s.Config = c
	}

	return err
}

func (c *JobConfig) parseVcapAppJSON() (err error) {
	c.VcapAppData = &VcapAppData{}
	c.VcapAppData, err = parseCfgJSON([]byte(c.VcapAppJSON), c.VcapAppData)
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

func CIVarsToMap(vars []*CIVar) map[string]string {
	if vars == nil {
		return nil
	}
	mapped := make(map[string]string)
	for _, v := range vars {
		mapped[v.Key] = v.Value
	}
	return mapped
}

func getJobConfig() (cfg *JobConfig, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error getting job config: %w", err)
		}
	}()

	cfg = (&JobConfig{}).parseEnv()
	if err = cfg.parseJobResponseFile(); err != nil {
		return nil, err
	}
	if err = cfg.parseVcapAppJSON(); err != nil {
		return nil, err
	}

	cfg.ContainerID = fmt.Sprintf(
		"glrw-p%v-c%v-j%v",
		cfg.ProjectID,
		cfg.ConcurrentProjectID,
		cfg.JobID,
	)

	cfg.Manifest = cloudgov.NewAppManifest(
		cfg.ContainerID,
		cfg.WorkerMemory,
		cfg.WorkerDiskSize,
	)

	cfg.Manifest.Env = CIVarsToMap(cfg.Variables)

	if cfg.Image != nil {
		img := cfg.Image.Name
		cfg.Manifest.Docker.Image = img

		// match images w/ docker domain, or no domain (i.e. docker by default)
		re := regexp.MustCompile(`^((registry-\d+|index)?\.?docker.io\/|[^.]*(:|$))`)

		if strings.Contains(img, "registry.gitlab.com") {
			cfg.Manifest.Docker.Username = cfg.CIRegistryUser
			cfg.Manifest.Docker.Password = cfg.CIRegistryPass
		} else if re.FindString(img) != "" {
			cfg.Manifest.Docker.Username = cfg.DockerHubUser
			cfg.Manifest.Docker.Password = cfg.DockerHubToken
		}

		var x []string
		for _, str := range append(cfg.Image.Entrypoint, cfg.Image.Command...) {
			str = strings.Trim(str, " ")
			if str != "" {
				x = append(x, str)
			}
		}
		cfg.Manifest.Process.Command = strings.Join(x, " ")
	}

	// TODO: still need to process service images when they get used
	for _, s := range cfg.Services {
		s.Manifest = cloudgov.NewAppManifest(
			fmt.Sprintf("%v-svc-%v", cfg.ContainerID, s.Alias),
			cfg.WorkerMemory,
			cfg.WorkerDiskSize,
		)
		s.Manifest.Env = CIVarsToMap(append(cfg.Variables, s.Variables...))
		var x []string
		for _, str := range append(s.Entrypoint, s.Command...) {
			str = strings.Trim(str, " ")
			if str != "" {
				x = append(x, str)
			}
		}
		s.Manifest.Process.Command = strings.Join(x, " ")
		fmt.Println(s.Manifest)
	}

	fmt.Println(cfg.Manifest)

	return cfg, nil
}
