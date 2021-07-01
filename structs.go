package main

type Package struct {
	Name		string		`json:"name,omitempty"`
	Epoch		string		`json:"epoch,omitempty"`
	Version		string		`json:"version,omitempty"`
	Release		string		`json:"release,omitempty"`
	Arch		string		`json:"arch,omitempty"`
	PkgManager	string		`json:"pkgmanager,omitempty"`
}

type Repository struct {
	Type		string		`json:"type,omitempty"`
	Name		string		`json:"name,omitempty"`
	Priority	int		`json:"priority,omitempty"`
	Url		string		`json:"url,omitempty"`
	Description	string		`json:"description,omitempty"`
}

type Report struct {
	Host		string		`json:"host,omitempty"`
	Tags		[]string	`json:"tags,omitempty"`
	Kernel		string		`json:"kernel,omitempty"`
	Arch		string		`json:"arch,omitempty"`
	Protocol	string		`json:"protocol,omitempty"`
	OS		string		`json:"os,omitempty"`
	Packages	[]Package
	Repos		[]Repository
	SecUpdates	[]Package
	BugUpdates	[]Package
	Reboot		string		`json:"reboot,omitempty"`
}

