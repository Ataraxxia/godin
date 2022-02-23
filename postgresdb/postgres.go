package postgresdb

import (
	"database/sql"
	"fmt"
	"time"

	rep "github.com/Ataraxxia/godin/report"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type DB struct {
	User          string
	Password      string
	DatabaseName  string
	ServerAddress string
	ServerPort    string
	MockDB        *sql.DB
}

type PostgreSQLConfiguration struct {
	SQLUser          string
	SQLPassword      string
	SQLDatabaseName  string
	SQLServerAddress string
	SQLPort          string
}

func checkTableExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(fmt.Sprintf("SELECT EXISTS ( SELECT FROM pg_tables WHERE  schemaname = 'public' AND tablename = '%s' );", name)).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (d DB) InitializeDatabase() error {
	db, err := d.getDatabaseHandle()

	if err != nil {
		return err
	}
	defer db.Close()

	if v, err := checkTableExists(db, "reports"); err == nil && v == false {
		log.Debug("Initialising PostgreSQL database")

		_, err = db.Exec(`CREATE TABLE reports(
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMP,
			hostname VARCHAR (255),
			report JSONB
		);`)

		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func (d DB) getDatabaseHandle() (*sql.DB, error) {

	if d.MockDB != nil {
		return d.MockDB, nil
	}

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", d.ServerAddress, d.ServerPort, d.User, d.Password, d.DatabaseName)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (d DB) SaveReport(r rep.Report, t time.Time) error {
	db, err := d.getDatabaseHandle()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO reports (timestamp, hostname, report) VALUES ($1,$2,$3)", t, r.HostInfo.Hostname, r)
	if err != nil {
		return err
	}

	return nil
}
