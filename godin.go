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
}

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
	payload, _ := ioutil.ReadAll(r.Body)
	log.Debug(string(payload))
	w.Write([]byte("Godin says OK"))
}