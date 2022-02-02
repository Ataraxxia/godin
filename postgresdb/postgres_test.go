package postgresdb

import (
	"database/sql/driver"
	"io/ioutil"
	"os"
	"testing"
	"time"

	rep "github.com/Ataraxxia/godin/report"
	"github.com/DATA-DOG/go-sqlmock"
)

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type AnyReport struct{}

func (a AnyReport) Match(v driver.Value) bool {
	_, ok := v.([]byte)
	return ok
}

type AnyString struct{}

func (a AnyString) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

func TestGetDatabaseHandle(t *testing.T) {

	// Testing mock connection
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := DB{
		User:          "",
		Password:      "",
		DatabaseName:  "",
		ServerAddress: "",
		MockDB:        db,
	}

	m, err := d.getDatabaseHandle()

	if err != nil {
		t.Errorf("Expected nil got %v", err)
	}

	if m != d.MockDB {
		t.Errorf("Incorrect pgsql mock pointer")
	}

	// Trying real connection
	d = DB{
		User:          "",
		Password:      "",
		DatabaseName:  "",
		ServerAddress: "",
		MockDB:        nil,
	}

	m, err = d.getDatabaseHandle()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestInitializeDatabase(t *testing.T) {

	// Database doesn't exist
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d := DB{
		User:          "",
		Password:      "",
		DatabaseName:  "",
		ServerAddress: "",
		MockDB:        db,
	}

	rows := sqlmock.NewRows([]string{"exists"}).
		AddRow(false)
	mock.ExpectQuery("^SELECT EXISTS \\( SELECT FROM pg_tables WHERE  schemaname = 'public' AND tablename = 'reports' \\)\\;$").WillReturnRows(rows)
	mock.ExpectExec("CREATE TABLE reports").WillReturnResult(sqlmock.NewResult(1, 1))

	err = d.InitializeDatabase()
	if err != nil {
		t.Errorf("Expected nil, got: %v", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	//Database exists
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	d = DB{
		User:          "",
		Password:      "",
		DatabaseName:  "",
		ServerAddress: "",
		MockDB:        db,
	}

	rows = sqlmock.NewRows([]string{"exists"}).
		AddRow(true)
	mock.ExpectQuery("^SELECT EXISTS \\( SELECT FROM pg_tables WHERE  schemaname = 'public' AND tablename = 'reports' \\)\\;$").WillReturnRows(rows)

	err = d.InitializeDatabase()
	if err != nil {
		t.Errorf("Expected nil, got: %v", err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSaveReport(t *testing.T) {
	const testdatapath string = "../testdata/"
	testtable := []struct {
		file     string
		expected bool
	}{
		{
			file:     "apt_report_ok_1.json",
			expected: true,
		},
	}

	d := DB{
		User:          "",
		Password:      "",
		DatabaseName:  "",
		ServerAddress: "",
		MockDB:        nil,
	}

	for _, tc := range testtable {
		f, err := os.Open(testdatapath + tc.file)
		if err != nil {
			t.Fatalf("Error opening test file %s", tc.file)
		}
		defer f.Close()

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		d.MockDB = db

		mock.ExpectExec("INSERT INTO reports").WithArgs(AnyTime{}, AnyString{}, AnyReport{}).WillReturnResult(sqlmock.NewResult(1, 1))

		content, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatalf("Error reading contents of golden file %v", tc.file)
		}

		var report rep.Report
		err = report.Scan(content)

		if err != nil {
			t.Fatalf("Error decoding JSON file: %v", err)
		}

		err = d.SaveReport(report, time.Now().UTC())
		if tc.expected {
			if err != nil {
				t.Errorf("%v Expected nil got %v", tc, err)
			}
		} else {
			if err == nil {
				t.Errorf("%v Expected error, got nil", tc)
			}
		}
	}
}
