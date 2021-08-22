package Report

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

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
	Protocol       string       `json:"protocol"`
	PackageManager string       `json:"package_manager"`
	Tags           string       `json:"tags"`
	Repositories   []Repository `json:"repositories,omitempty"`
	Packages       []Package    `json:"packages"`
}

// Make the Attrs struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (r Report) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Make the Attrs struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (r *Report) Scan(value interface{}) error {
	b, err := value.([]byte)
	if !err {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}
