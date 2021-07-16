package main

import (
    "database/sql"
//    "database/sql/driver"
//   "encoding/json"
//    "errors"
    "fmt"

    _ "github.com/lib/pq"
    log "github.com/sirupsen/logrus"
)

func checkTableExists(db *sql.DB, name string) bool {
	var exists bool
	err := db.QueryRow(fmt.Sprintf("SELECT EXISTS ( SELECT FROM pg_tables WHERE  schemaname = 'public' AND tablename = '%s' );", name)).Scan(&exists)
	if err != nil {
		log.Error(err)
	}
	return exists
}

func initDB() error {
	db, err := sql.Open("postgres", "postgres://godin:password@localhost/godin") //todo parametrize
	if err != nil {
		log.Fatal(err)
	}

	if checkTableExists(db, "reports") == false {
		log.Debug("Initialising DB")

		_, err = db.Exec(`CREATE TABLE reports(
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMP,
			report JSONB
		);`)

		if err != nil {
			log.Error(err)
		}
	}

	log.Debug("DB init done")

	return nil
}