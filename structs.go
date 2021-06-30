package main

type Package struct {
	Name		string
	Epoch		string
	Version		string
	Release		string
	Arch		string
	PkgManager	string
}

type Repository struct {
	Type		string
	Name		string
	Priority	int
	Url		string
	Description	string
}

type Report struct {
	Host		string
	Tags		[]string
	Kernel		string
	Arch		string
	Protocol	string
	OS		string
	Packages	[]Package
	Repos		[]Repository
	SecUpdates	[]Package
	BugUpdates	[]Package
	Reboot		string
}

