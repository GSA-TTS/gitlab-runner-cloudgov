package cloudgov

type AppManifest struct {
	Name      string // i.e., container ID
	Env       map[string]string
	NoRoute   bool
	Docker    AppManifestDocker
	Process   *AppManifestProcess
	OrgName   string
	SpaceName string
}

type AppManifestDocker struct {
	Image    string
	Username string
	Password string
}

type AppManifestProcess struct {
	Command         string // Entrypoint + Cmd
	DiskQuota       string
	Memory          string
	HealthCheckType string
}

func NewAppManifest(id string, org string, space string, memory string, disk string) AppManifest {
	return AppManifest{
		Name:      id,
		OrgName:   org,
		SpaceName: space,
		NoRoute:   true,
		Process: &AppManifestProcess{
			Memory:          memory,
			DiskQuota:       disk,
			HealthCheckType: "process",
		},
	}
}
