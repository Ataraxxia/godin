package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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
	SQLUser         string
	SQLPassword     string
	SQLDatabaseName string
	SQLServerAddr   string
}

var (
	BuildVersion string = ""
	BuildTime    string = ""

	config *configuration
	//	db     postgresdb.DB
	db Database

	configFilePathPtr = flag.String("config", "/etc/godin/settings.json", "Path to configuration file")
	versionPtr        = flag.Bool("version", false, "Display version and exit")
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

	db = postgresdb.DB{
		User:          config.SQLUser,
		Password:      config.SQLPassword,
		DatabaseName:  config.SQLDatabaseName,
		ServerAddress: config.SQLServerAddr,
		MockDB:        nil,
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
			log.Printf("Syntax error at byte offset %d\n", e.Offset)
		}

		w.Write([]byte("Godin says Json decoding error\n"))
		return
	}

	t := time.Now().UTC()
	err = db.SaveReport(report, t)
	if err != nil {
		log.Error(err)
	}

	w.Write([]byte("Godin says OK\n"))
}
