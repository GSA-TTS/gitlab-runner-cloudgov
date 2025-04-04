package drive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GSA-TTS/gitlab-runner-cloudgov/runner/cfd/cloudgov"
	"github.com/google/go-cmp/cmp"
)

// TODO: can replace this and the other tests with one nice table test,
// but I think it'll be easier to manage if parsing a file and want to
// think about that later.
func Test_GetJobConfig(t *testing.T) {
	cfgWant := &JobConfig{
		JobResponse:      JobResponse{},
		CIRegistryUser:   "foo",
		CIRegistryPass:   "bar",
		DockerHubUser:    "foo",
		DockerHubToken:   "1234",
		WorkerMemory:     "1024M",
		WorkerDiskSize:   "1024M",
		JobResponseFile:  "",
		VcapAppJSON:      "",
		VcapServicesData: VcapServicesData{},
		ContainerID:      "glrw-p-c-j",
		Manifest: &cloudgov.AppManifest{
			Name:    "glrw-p-c-j",
			NoRoute: true,
			Process: cloudgov.AppManifestProcess{
				DiskQuota: "1024M", Memory: "1024M", HealthCheckType: "process",
			},
		},
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
	parsedCfg, err := getJobConfig()
	if err != nil {
		t.Error(err)
		return
	}
	if diff := cmp.Diff(parsedCfg, cfgWant); diff != "" {
		t.Error(diff)
	}
}

func Test_parseJobResponseFile(t *testing.T) {
	t.Setenv("JOB_RESPONSE_FILE", "./testdata/sample_job_response.json")

	wanted := JobResponse{
		Image: Image{
			Name:       "ubuntu:jammy",
			Command:    []string{"a", "b", "c"},
			Entrypoint: []string{"d", "e", "f"},
		},
		Services: []*Service{{
			Image: Image{
				Name:       "postgres:wormy",
				Alias:      "my-pg-service",
				Command:    []string{"g", "h", "i"},
				Entrypoint: []string{"j", "k", "l"},
			},
			Variables: []CIVar{{Key: "bazz", Value: "buzz"}},
			Manifest: &cloudgov.AppManifest{
				Name:    "glrw-p-c-j-svc-my-pg-service",
				Env:     map[string]string{"bazz": "buzz", "foo": "bar"},
				NoRoute: true,
				Docker:  cloudgov.AppManifestDocker{Image: "postgres:wormy"},
				Process: cloudgov.AppManifestProcess{Command: "j k l g h i", HealthCheckType: "process"},
			},
			Config: &JobConfig{
				ContainerID:      "glrw-p-c-j",
				JobResponseFile:  "./testdata/sample_job_response.json",
				VcapServicesData: VcapServicesData{},
				Manifest: &cloudgov.AppManifest{
					Name:    "glrw-p-c-j",
					Env:     map[string]string{"foo": "bar"},
					NoRoute: true,
					Docker:  cloudgov.AppManifestDocker{Image: "ubuntu:jammy"},
					Process: cloudgov.AppManifestProcess{Command: "d e f a b c", HealthCheckType: "process"},
				},
			},
		}},
		Variables: []CIVar{{Key: "foo", Value: "bar"}},
	}

	// here to complete the cicurular reference from services back to cfg
	wanted.Services[0].Config.JobResponse = wanted

	cfg, err := getJobConfig()
	if err != nil {
		t.Error(err)
		return
	}
	if diff := cmp.Diff(cfg.JobResponse, wanted); diff != "" {
		t.Errorf("msmatch (-got +want):\n%s", diff)
	}
}

func Test_parseVcapAppJSON(t *testing.T) {
	sample := `{"cf_api":"https://api.fr.cloud.gov","limits":{"fds":16384,"mem":128,"disk":1024},"application_name":"gitlab-runner","application_uris":[],"name":"gitlab-runner","space_name":"zjr-gl-test","space_id":"8969a4b6-01aa-431d-9790-77cc4c47e3e7","organization_id":"f0a46189-6f64-43fb-99c3-0719cf9ee255","organization_name":"gsa-tts-devtools-prototyping","uris":[],"process_id":"e905fbb9-aea0-44aa-ba10-f76aed1668d1","process_type":"web","application_id":"e905fbb9-aea0-44aa-ba10-f76aed1668d1","version":"f115779a-17a3-4700-9941-aae3fe81a4c8","application_version":"f115779a-17a3-4700-9941-aae3fe81a4c8"}`
	t.Setenv("VCAP_APPLICATION", sample)

	wanted := VcapAppData{
		CFApi:     "https://api.fr.cloud.gov",
		OrgName:   "gsa-tts-devtools-prototyping",
		SpaceId:   "8969a4b6-01aa-431d-9790-77cc4c47e3e7",
		SpaceName: "zjr-gl-test",
	}

	cfg, err := getJobConfig()
	if err != nil {
		t.Error(err)
		return
	}

	if diff := cmp.Diff(cfg.VcapAppData, wanted); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func Test_parseVcapServicesJSON(t *testing.T) {
	sample := `{"s3":[{"label":"s3","provider":null,"plan":"basic-sandbox","name":"glr-dependency-cache","tags":["AWS","S3","object-storage","terraform-cloudgov-managed"],"instance_guid":"d1541026","instance_name":"glr-dependency-cache","binding_guid":"9f316c56","binding_name":null,"credentials":{"uri":"s3://goooo:booo@s3-fips.us-gov-west-1.aaws.com/cg-d1541026","insecure_skip_verify":false,"access_key_id":"jjjjj","secret_access_key":"ssssssss","region":"us-gov-west-1","bucket":"cg-d1541026","endpoint":"s3-fips.us-gov-west-1.amazonaws.com","fips_endpoint":"s3-fips.us-gov-west-1.amazonaws.com","additional_buckets":[]},"syslog_drain_url":null,"volume_mounts":[]}],"user-provided":[{"label":"user-provided","name":"glr-egress-proxy-credentials","tags":[],"instance_guid":"608e3f73","instance_name":"glr-egress-proxy-credentials","binding_guid":"7530ea7b","binding_name":null,"credentials":{"cred_string":"018052ba:ukgZ","domain":"egress-proxy.apps.internal","http_port":8080,"http_uri":"http://018052b:ukHK@egress-proxy.apps.internal:8080","https_uri":"https://018052ba:ukHK1@egress-proxy.apps.internal:61443"},"syslog_drain_url":null,"volume_mounts":[]}]}`
	t.Setenv("VCAP_SERVICES", sample)
	t.Setenv("PROXY_CREDENTIAL_INSTANCE", "glr-egress-proxy-credentials")

	dir, err := os.MkdirTemp("", "temp_auth_files")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	authFile := filepath.Join(dir, "ssh_proxy.auth")
	t.Setenv("PROXY_AUTH_FILE", authFile)

	credStringWanted := "018052ba:ukgZ"

	wantedServices := VcapServicesData{
		"s3": []VcapServiceInstance{{
			Name: "glr-dependency-cache",
		}},
		"user-provided": []VcapServiceInstance{{
			Name: "glr-egress-proxy-credentials",
			Credentials: VcapServiceCredentials{
				Domain:     "egress-proxy.apps.internal",
				HTTPPort:   8080,
				HTTPURI:    "http://018052b:ukHK@egress-proxy.apps.internal:8080",
				HTTPSURI:   "https://018052ba:ukHK1@egress-proxy.apps.internal:61443",
				CredString: credStringWanted,
			},
		}},
	}

	wantedEgressConfig := EgressProxyConfig{
		ProxyHostHTTP:  "http://018052b:ukHK@egress-proxy.apps.internal:8080",
		ProxyHostHTTPS: "https://018052ba:ukHK1@egress-proxy.apps.internal:61443",
		ProxyHostSSH:   "egress-proxy.apps.internal",
		ProxyPortSSH:   8080,
		ProxyAuthFile:  authFile,
	}

	cfg, err := getJobConfig()
	if err != nil {
		t.Error(err)
		return
	}

	if diff := cmp.Diff(cfg.VcapServicesData, wantedServices); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}

	if diff := cmp.Diff(cfg.VcapServicesData["user-provided"][0], wantedServices["user-provided"][0]); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}

	if diff := cmp.Diff(cfg.EgressProxyConfig, wantedEgressConfig); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}

	credString, err := os.ReadFile(cfg.ProxyAuthFile)
	if err != nil {
		t.Fatalf("error reading ProxyAuthFile: %v", err)
	}
	if diff := cmp.Diff(string(credString), credStringWanted); diff != "" {
		t.Fatalf("mismatch (-got +want):\n%s", diff)
	}
}
