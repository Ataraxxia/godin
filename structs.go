package main

type Package struct {
	Name		string
	Epoch		string
	Version		string
	Release		string
	Arch		string
	PkgManager	string
}

type Report struct {
	Host		string
	Tags		[]string
	Kernel		string
	Arch		string
	Protocol	string
	OS		string
	Packages	[]Package
	Repos		[]string
	SecUpdates	[]string
	BugUpdates	[]string
	Reboot		string
}

