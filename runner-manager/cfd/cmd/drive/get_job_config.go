package drive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
)

type JobConfig struct {
	JobResponse            // Parsed JSON from JobResponseFile
	JobResponseFile string `env:"JOB_RESPONSE_FILE"`

	VcapAppData
	VcapAppJSON string `env:"VCAP_APPLICATION"`

	VcapServicesData
	VcapServicesJSON  string `env:"VCAP_SERVICES"`
	EgressServiceName string `env:"PROXY_CREDENTIAL_INSTANCE"`

	EgressProxyConfig

	Manifest *cloudgov.AppManifest

	// We combine the following to make the container ID.
	// Some are available in JOB_RESPONSE_FILE, but several are only found
	// within `.variables` so we pull all from ENV for consistency.
	ContainerID         string
	JobID               string `env:"CUSTOM_ENV_CI_JOB_ID"`
	RunnerID            string `env:"CUSTOM_ENV_CI_RUNNER_ID"`
	ProjectID           string `env:"CUSTOM_ENV_CI_PROJECT_ID"`
	ConcurrentProjectID string `env:"CUSTOM_ENV_CI_CONCURRENT_PROJECT_ID"`

	// TODO: #95 - we might want to grab/store these differently
	CIRegistryUser string `env:"CUSTOM_ENV_CI_REGISTRY_USER"`
	CIRegistryPass string `env:"CUSTOM_ENV_CI_REGISTRY_PASSWORD"`
	DockerHubUser  string `env:"DOCKER_HUB_USER"`
	DockerHubToken string `env:"DOCKER_HUB_TOKEN"`

	WorkerMemory   string `env:"WORKER_MEMORY"`
	WorkerDiskSize string `env:"WORKER_DISK_SIZE"`
}

type JobResponse struct {
	Image     Image
	Variables []CIVar
	Services  []*Service
}
type Image struct {
	Name       string
	Alias      string
	Command    []string
	Entrypoint []string
}
type Service struct {
	Image
	Variables []CIVar
	Manifest  *cloudgov.AppManifest
	Config    *JobConfig
}
type CIVar struct {
	Key   string
	Value string
}

type VcapAppData struct {
	CFApi     string `json:"cf_api"`
	OrgID     string `json:"org_id"`
	OrgName   string `json:"organization_name"`
	SpaceId   string `json:"space_id"`
	SpaceName string `json:"space_name"`
}

type (
	VcapServicesData    map[string][]VcapServiceInstance
	VcapServiceInstance struct {
		Name        string
		Credentials VcapServiceCredentials
	}
)

type VcapServiceCredentials struct {
	Domain     string `json:"domain"`
	HTTPPort   int    `json:"http_port"`
	HTTPURI    string `json:"http_uri"`
	HTTPSURI   string `json:"https_uri"`
	CredString string `json:"cred_string"`
}

type EgressProxyConfig struct {
	ProxyHostHTTP  string
	ProxyHostHTTPS string
	ProxyHostSSH   string
	ProxyPortSSH   int
	ProxyAuthFile  string
}

func parseCfgJSON[R any](j []byte, r *R) (*R, error) {
	if len(j) < 1 {
		return r, nil
	}
	if err := json.Unmarshal(j, r); err != nil {
		return nil, fmt.Errorf("error parsing %t: %w", r, err)
	}
	return r, nil
}

func (c *JobConfig) parseJobResponseFile() (err error) {
	ref := &JobResponse{}

	if c.JobResponseFile == "" {
		return nil
	}

	j, err := os.ReadFile(c.JobResponseFile)
	if err != nil {
		return fmt.Errorf("error reading JobResponseFile: %w", err)
	}

	ref, err = parseCfgJSON(j, ref)
	if err != nil {
		return err
	}

	c.JobResponse = *ref

	for _, s := range c.Services {
		s.Config = c
	}

	return err
}

func (c *JobConfig) parseVcapAppJSON() (err error) {
	ref := &VcapAppData{}
	ref, err = parseCfgJSON([]byte(c.VcapAppJSON), ref)
	c.VcapAppData = *ref
	return err
}

func (c *JobConfig) parseVcapServicesJSON() (err error) {
	ref := &map[string][]VcapServiceInstance{}
	ref, err = parseCfgJSON([]byte(c.VcapServicesJSON), ref)
	c.VcapServicesData = *ref
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

func (cfg *JobConfig) makeManifest(id string) *cloudgov.AppManifest {
	return &cloudgov.AppManifest{
		Name:      id,
		OrgName:   cfg.OrgName,
		SpaceName: cfg.SpaceName,
		NoRoute:   true,
		Process: cloudgov.AppManifestProcess{
			Memory:          cfg.WorkerMemory,
			DiskQuota:       cfg.WorkerDiskSize,
			HealthCheckType: "process",
		},
	}
}

func (cfg *JobConfig) ciVarsToMap(vars []CIVar, m *cloudgov.AppManifest) {
	if vars == nil {
		return
	}
	m.Env = make(map[string]string)
	for _, v := range vars {
		m.Env[v.Key] = v.Value
	}
}

func (cfg *JobConfig) processImage(img Image, m *cloudgov.AppManifest) {
	if img.Name != "" {
		m.Docker.Image = img.Name

		// match images w/ docker domain, or no domain (i.e. docker by default)
		re := regexp.MustCompile(`^((registry-\d+|index)?\.?docker\.io\/|[^.]*(:|$))`)

		// TODO: #95
		if strings.Contains(img.Name, "registry.gitlab.com") {
			m.Docker.Username = cfg.CIRegistryUser
			m.Docker.Password = cfg.CIRegistryPass
		} else if re.FindString(img.Name) != "" {
			m.Docker.Username = cfg.DockerHubUser
			m.Docker.Password = cfg.DockerHubToken
		}

		var x []string
		for _, str := range append(img.Entrypoint, img.Command...) {
			str = strings.Trim(str, " ")
			if str != "" {
				x = append(x, str)
			}
		}
		m.Process.Command = strings.Join(x, " ")
	}
}

func (cfg *JobConfig) processEgressProxyCfg() (err error) {
	defer (func() {
		if err != nil {
			err = fmt.Errorf("error processEgressProxyCfg: %w", err)
		}
	})()

	userServices := cfg.VcapServicesData["user-provided"]
	if len(userServices) < 1 {
		return nil
	}

	egressIdx := slices.IndexFunc(userServices, func(vsi VcapServiceInstance) bool {
		return vsi.Name == cfg.EgressServiceName
	})
	if egressIdx < 0 {
		return nil
	}

	esc := userServices[egressIdx].Credentials

	cfg.EgressProxyConfig = EgressProxyConfig{
		ProxyHostHTTP:  esc.HTTPURI,
		ProxyHostHTTPS: esc.HTTPSURI,
		ProxyHostSSH:   esc.Domain,
		ProxyPortSSH:   esc.HTTPPort,
		ProxyAuthFile:  os.Getenv("PROXY_AUTH_FILE"),
	}

	if cfg.ProxyAuthFile == "" {
		cfg.ProxyAuthFile = "/home/vcap/app/ssh_proxy.auth"
	}

	return os.WriteFile(cfg.ProxyAuthFile, []byte(esc.CredString), 0600)
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
	if err = cfg.parseVcapServicesJSON(); err != nil {
		return nil, err
	}
	if err = cfg.processEgressProxyCfg(); err != nil {
		return nil, err
	}

	cfg.ContainerID = fmt.Sprintf(
		"glrw-p%v-c%v-j%v",
		cfg.ProjectID,
		cfg.ConcurrentProjectID,
		cfg.JobID,
	)

	cfg.Manifest = cfg.makeManifest(cfg.ContainerID)
	cfg.ciVarsToMap(cfg.Variables, cfg.Manifest)
	cfg.processImage(cfg.Image, cfg.Manifest)

	for _, s := range cfg.Services {
		serviceId := fmt.Sprintf("%v-svc-%v", cfg.ContainerID, s.Alias)
		s.Manifest = cfg.makeManifest(serviceId)
		cfg.ciVarsToMap(append(cfg.Variables, s.Variables...), s.Manifest)
		cfg.processImage(s.Image, s.Manifest)
	}

	return cfg, nil
}
