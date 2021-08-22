package Report

type Package struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Repository   string `json:"repository"`
	Upgrade      string `json:"upgrade,omitempty"`
}

type Repository struct {
	RepositoryAlias   string `json:"repository_alias"`
	RepositoryID      string `json:"repository_id"`
	RepositoryName    string `json:"repository_name"`
	RepositoryBaseURL string `json:"repository_baseurl"`
}

type HostInfo struct {
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Hostname     string `json:"hostname"`
}

type Report struct {
	HostInfo       HostInfo     `json:"host_info"`
	RepositoryType string       `json:"repo_type"`
	PackageManager string       `json:"package_manager"`
	Tags           string       `json:"tags"`
	Repositories   []Repository `json:"repositories,omitempty"`
	Packages       []Package    `json:"packages"`
}
