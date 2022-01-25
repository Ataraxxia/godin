package postgresdb

import (
	"database/sql"
	"time"

	"fmt"

	rep "github.com/Ataraxxia/godin/Report"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	User          string
	Password      string
	DatabaseName  string
	ServerAddress string
}

func checkTableExists(db *sql.DB, name string) bool {
	var exists bool
	err := db.QueryRow(fmt.Sprintf("SELECT EXISTS ( SELECT FROM pg_tables WHERE  schemaname = 'public' AND tablename = '%s' );", name)).Scan(&exists)
	if err != nil {
		log.Error(err)
	}
	return exists
}

func (d DB) getConnString() {
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", d.User, d.Password, d.ServerAddress, d.DatabaseName)
	return connString
}

func (d DB) InitDB() error {
	db, err := sql.Open("postgres", d.getConnString())
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	if checkTableExists(db, "reports") == false {
		log.Debug("Initialising DB")

		_, err = db.Exec(`CREATE TABLE reports(
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMP,
			hostname VARCHAR (255),
			report JSONB
		);`)

		if err != nil {
			log.Error(err)
		}
	}

	log.Debug("DB init done")

	return nil
}

func (d DB) SaveReport(r rep.Report) error {
	db, err := sql.Open("postgres", d.getConnString())
	if err != nil {
		return err
	}
	defer db.Close()

	reportTime := time.Now().UTC()
	_, err = db.Exec("INSERT INTO reports (timestamp, hostname, report) VALUES ($1,$2,$3)", reportTime, r.HostInfo.Hostname, r)
	if err != nil {
		return err
	}

	return nil
}
