package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Ataraxxia/godin/mongodb"
	"github.com/Ataraxxia/godin/postgresdb"

	rep "github.com/Ataraxxia/godin/report"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type configuration struct {
	Address         string
	Port            string
	LogLevel        string
	DatabaseBackend string
	PostgreSQL      postgresdb.PostgreSQLConfiguration `json:"PostgreSQL,omitempty"`
	MongoDB         mongodb.MongoDBConfiguration       `json:"PostgreSQL,omitempty"`
}

var (
	BuildVersion string = ""
	BuildTime    string = ""

	config *configuration
	db     Database

	configFilePathPtr = flag.String("config", "/etc/godin/settings.json", "Path to configuration file")
	versionPtr        = flag.Bool("version", false, "Display version and exit")
)

const (
	MSG_OK                = "Godin OK"
	MSG_SERVER_ERROR      = "Server side error"
	MSG_JSON_ERROR        = "Error decoding JSON"
	MSG_JSON_HEADER_ERROR = "Content-Type header is not application/json"
	MSG_NO_BACKEND_ERROR  = "Database backend not specified or not supported"
)

func loadConfig(fpath string) error {
	f, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(f), &config)
	if err != nil {
		return err
	}
	switch loglevel := strings.ToLower(config.LogLevel); loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	default:
		log.SetLevel(log.InfoLevel)
	}

	switch strings.ToLower(config.DatabaseBackend) {
	case "postgresql":
		db = postgresdb.DB{
			User:          config.PostgreSQL.SQLUser,
			Password:      config.PostgreSQL.SQLPassword,
			DatabaseName:  config.PostgreSQL.SQLDatabaseName,
			ServerAddress: config.PostgreSQL.SQLServerAddress,
			ServerPort:    config.PostgreSQL.SQLPort,
			MockDB:        nil,
		}
	case "mongodb":
		db = mongodb.DB{
			User:         config.MongoDB.User,
			Password:     config.MongoDB.Password,
			DatabaseName: config.MongoDB.DatabaseName,
			URI:          config.MongoDB.URI,
		}
	default:
		log.Fatal(MSG_NO_BACKEND_ERROR)
	}

	return nil
}

func main() {
	var err error
	flag.Parse()

	if *versionPtr {
		fmt.Printf("Godin Server v%s\n", BuildVersion)
		return
	}

	fpath := fmt.Sprintf(*configFilePathPtr)
	if err = loadConfig(fpath); err != nil {
		log.Fatal(err)
	}

	err = db.InitializeDatabase()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", getDefaultPage)
	r.HandleFunc("/reports/upload", uploadReport).Methods("POST")

	log.Infof("Starting server %s:%s", config.Address, config.Port)
	err = http.ListenAndServe(config.Address+":"+config.Port, r)
	log.Info(err)
}

func getDefaultPage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, MSG_OK, http.StatusOK)
}

func uploadReport(w http.ResponseWriter, r *http.Request) {
	log.Debug("Getting new report")

	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			http.Error(w, MSG_JSON_HEADER_ERROR, http.StatusUnsupportedMediaType)
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
			log.Printf("Syntax error at byte offset %d\n", e.Offset)
		}
		http.Error(w, MSG_JSON_ERROR, http.StatusBadRequest)
		return
	}

	log.Debugf("Saving report from %s", report.HostInfo.Hostname)

	t := time.Now().UTC()
	err = db.SaveReport(report, t)
	if err != nil {
		log.Error(err)
		http.Error(w, MSG_SERVER_ERROR, http.StatusInternalServerError)
	} else {
		http.Error(w, MSG_OK, http.StatusOK)
	}
	return
}
