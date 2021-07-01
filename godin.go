package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"flag"
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

	initDB()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", getDefaultPage)
	r.HandleFunc("/reports/upload/", uploadReport).Methods("POST")

	log.Infof("Starting server %s:%s", config.Address, config.Port)
	err = http.ListenAndServe(config.Address+":"+config.Port, r)
	log.Info(err)

}

func getDefaultPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Godin"))
}


func uploadReport(w http.ResponseWriter, r *http.Request) {
	log.Debug("Getting new report")
	if err := r.ParseMultipartForm(maxMemoryBytes); err != nil {
		log.Error(err)
	}

	reportName, err := saveReport(r.Form)
	if err != nil {
		log.Error(err)
	}

	log.Debugf("Report %s saved", reportName)

	w.Write([]byte("Godin says OK"))
}


