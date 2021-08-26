package main

import (
	"encoding/json"
	"flag"
	"github.com/Ataraxxia/godin/postgresdb"
	"io/ioutil"
	"net/http"
	"strings"

	rep "github.com/Ataraxxia/godin/Report"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type configuration struct {
	Address         string
	Port            string
	LogLevel        string
	DataPath        string
	SQLUser         string
	SQLPassword     string
	SQLDatabaseName string
	SQLServerAddr   string
}

const (
	maxMemoryBytes = 10485760 // 10MiB
)

var (
	config *configuration
	db     postgresdb.DB
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

	db = postgresdb.DB{
		User:         config.SQLUser,
		Password:     config.SQLPassword,
		DatabaseName: config.SQLDatabaseName,
		ServerAddres: config.SQLServerAddr,
	}

	err = db.InitDB()
	if err != nil {
		return
	}

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

	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var report rep.Report
	err := dec.Decode(&report)
	if err != nil {
		log.Error(err)

		if e, err := err.(*json.SyntaxError); err {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}

		w.Write([]byte("Godin says Json decoding error"))
		return
	}

	err = db.SaveReport(report)
	if err != nil {
		log.Error(err)
	}

	w.Write([]byte("Godin says OK\n"))
}
