package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"flag"
	"os"
	"time"
	"fmt"
	"strings"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type configuration struct {
	Address		string
	Port		string
	LogLevel	string
	DataPath	string
}

const (
	maxMemoryBytes = 10485760 // 10MiB
)

var (
	config * configuration
)

func loadConfig() {
	flag.Parse()

	f, err := ioutil.ReadFile("/etc/godin/settings.json")
	if err != nil {
		log.Fatal("Coulnd't find /etc/godin/settings.json")
	}

	err = json.Unmarshal([]byte(f), &config)
	switch loglevel := strings.ToLower(config.LogLevel); loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Remove trailing slash
	if strings.HasSuffix(config.DataPath, "/") {
		tmp := config.DataPath
		config.DataPath = tmp[:len(tmp)-1]
	}

}

func main() {
	var err error
	loadConfig()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", getDefaultPage)
	r.HandleFunc("/reports/upload/", saveReport).Methods("POST")

	log.Infof("Starting server %s:%s", config.Address, config.Port)
	err = http.ListenAndServe(config.Address+":"+config.Port, r)
	log.Info(err)

}

func getDefaultPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Godin"))
}


func saveReport(w http.ResponseWriter, r *http.Request) {
	log.Debug("Getting new report")
	if err := r.ParseMultipartForm(maxMemoryBytes); err != nil {
		log.Error(err)
	}

	var report Report
	report.Host = r.Form["host"][0]
	report.Tags = strings.Split(r.Form["tags"][0], " ")
	report.Kernel = r.Form["kernel"][0]
	report.Arch = r.Form["arch"][0]
	report.Protocol = r.Form["protocol"][0]
	report.OS = r.Form["os"][0]
	report.Repos = strings.Split(r.Form["repos"][0], "\n")
	report.SecUpdates = strings.Split(r.Form["sec_updates"][0], "\n")
	report.BugUpdates = strings.Split(r.Form["bug_updates"][0], "\n")
	report.Reboot = r.Form["reboot"][0]


	packages := strings.Split(r.Form["packages"][0], "\n")
	for _, pack := range packages {
		pack = strings.ReplaceAll(pack, "'", "")
		p := strings.Split(pack, " ")
		if len(p) < 6 {
			continue
		}

		var pkg Package
		pkg.Name = p[0]
		pkg.Epoch = p[1]
		pkg.Version = p[2]
		pkg.Release = p[3]
		pkg.Arch = p[4]
		pkg.PkgManager = p[5]

		report.Packages = append(report.Packages, pkg)
	}

	bytes, err := json.Marshal(report)
	if err != nil {
		log.Error(err)
	}

	dt := strings.ReplaceAll(time.Now().Format("01-02-2006 15:04:05"), " ", "_")

	fpath := fmt.Sprintf("%s/%s_%s.json", config.DataPath, report.Host, dt)
	ioutil.WriteFile(fpath, bytes, os.ModePerm)

	log.Debug("Report saved")

	w.Write([]byte("Godin says OK"))
}


