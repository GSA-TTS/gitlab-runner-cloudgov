package drive

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TODO: can replace this and the other tests with one nice table test,
// but I think it'll be easier to manage if parsing a file and want to
// think about that later.
func Test_GetJobConfig(t *testing.T) {
	cfgWant := &JobConfig{
		CIRegistryUser:  "foo",
		CIRegistryPass:  "bar",
		DockerHubUser:   "foo",
		DockerHubToken:  "1234",
		WorkerMemory:    "1024M",
		WorkerDiskSize:  "1024M",
		JobResponseFile: "",
		VcapAppJSON:     "",
	}

	envVarsToSet := map[string]string{
		"JOB_RESPONSE_FILE":               cfgWant.JobResponseFile,
		"CUSTOM_ENV_CI_REGISTRY_USER":     cfgWant.CIRegistryUser,
		"CUSTOM_ENV_CI_REGISTRY_PASSWORD": cfgWant.CIRegistryPass,
		"DOCKER_HUB_TOKEN":                cfgWant.DockerHubToken,
		"DOCKER_HUB_USER":                 cfgWant.DockerHubUser,
		"WORKER_MEMORY":                   cfgWant.WorkerMemory,
		"WORKER_DISK_SIZE":                cfgWant.WorkerDiskSize,
		"VCAP_APPLICATION":                cfgWant.VcapAppJSON,
	}

	for k, v := range envVarsToSet {
		t.Setenv(k, v)
	}
	parsedCfg := GetJobConfig()
	if diff := cmp.Diff(cfgWant, parsedCfg); diff != "" {
		t.Error(diff)
	}
}

func Test_parseJobResponseFile(t *testing.T) {
	t.Setenv("JOB_RESPONSE_FILE", "./testdata/sample_job_response.json")

	wanted := &JobResponse{
		Image: &Image{
			Command:    []string{"a", "b", "c"},
			Entrypoint: []string{"d", "e", "f"},
		},
		Services: []*Service{{
			Image: &Image{
				Name:       "postgres:wormy",
				Alias:      "my-pg-service",
				Command:    []string{"g", "h", "i"},
				Entrypoint: []string{"j", "k", "l"},
			},
			Variables: []*CIVar{{Key: "bazz", Value: "buzz"}},
		}},
		Variables: []*CIVar{{Key: "foo", Value: "bar"}},
	}

	cfg := GetJobConfig()

	if diff := cmp.Diff(wanted, cfg.JobResponse); diff != "" {
		t.Error(diff)
	}
}

func Test_parseVcapAppJSON(t *testing.T) {
	sample := `{"cf_api":"https://api.fr.cloud.gov","limits":{"fds":16384,"mem":128,"disk":1024},"application_name":"gitlab-runner","application_uris":[],"name":"gitlab-runner","space_name":"zjr-gl-test","space_id":"8969a4b6-01aa-431d-9790-77cc4c47e3e7","organization_id":"f0a46189-6f64-43fb-99c3-0719cf9ee255","organization_name":"gsa-tts-devtools-prototyping","uris":[],"process_id":"e905fbb9-aea0-44aa-ba10-f76aed1668d1","process_type":"web","application_id":"e905fbb9-aea0-44aa-ba10-f76aed1668d1","version":"f115779a-17a3-4700-9941-aae3fe81a4c8","application_version":"f115779a-17a3-4700-9941-aae3fe81a4c8"}`
	t.Setenv("VCAP_APPLICATION", sample)

	wanted := &VcapAppData{
		OrgName:   "gsa-tts-devtools-prototyping",
		SpaceName: "zjr-gl-test",
	}

	cfg := GetJobConfig()

	if diff := cmp.Diff(wanted, cfg.VcapAppData); diff != "" {
		t.Error(diff)
	}
}
