package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"testing"

	rep "github.com/Ataraxxia/godin/report"
	log "github.com/sirupsen/logrus"
)

type MockDB struct {
	initializeDatabaseErr error
	saveReportErr         error
}

func (m MockDB) InitializeDatabase() error {
	return m.initializeDatabaseErr
}

func (m MockDB) SaveReport(r rep.Report, t time.Time) error {
	return m.initializeDatabaseErr
}

func TestGetDefaultPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	getDefaultPage(w, req)

	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != MSG_OK+"\n" {
		t.Errorf("expected %s got %v", MSG_OK, string(data))
	}
}

func TestUploadReport(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	const testdatapath string = "testdata/godin/"
	testtable := []struct {
		file     string
		headers  []map[string]string
		db       Database
		msg      string
		expected error
	}{
		// Everything OK
		{
			file: "apt_report_ok_1.json",
			headers: []map[string]string{
				{"Content-Type": "application/json"},
			},
			db: MockDB{
				initializeDatabaseErr: nil,
				saveReportErr:         nil,
			},
			msg:      MSG_OK,
			expected: nil,
		},

		// Malformed JSON
		{
			file: "apt_report_bad_1.json",
			headers: []map[string]string{
				{"Content-Type": "application/json"},
			},
			db: MockDB{
				initializeDatabaseErr: nil,
				saveReportErr:         nil,
			},
			msg:      MSG_JSON_ERROR,
			expected: nil,
		},

		// Wrong content header
		{
			file: "apt_report_ok_1.json",
			headers: []map[string]string{
				{"Content-Type": "application/xml"},
			},
			db: MockDB{
				initializeDatabaseErr: nil,
				saveReportErr:         nil,
			},
			msg:      MSG_JSON_HEADER_ERROR,
			expected: nil,
		},
	}

	for _, tc := range testtable {
		f, err := os.Open(testdatapath + tc.file)
		if err != nil {
			t.Fatalf("Could not open test file %s", tc.file)
		}

		db = tc.db

		req := httptest.NewRequest(http.MethodPost, "/reports/upload", f)
		for _, h := range tc.headers {
			for key, val := range h {
				req.Header.Add(key, val)
			}
		}
		w := httptest.NewRecorder()

		uploadReport(w, req)

		res := w.Result()
		defer res.Body.Close()

		data, err := ioutil.ReadAll(res.Body)

		if err != tc.expected {
			t.Errorf("expected error to be nil, got %v", err)
		}

		if string(data) != tc.msg+"\n" {
			t.Errorf("expected %s got %v", tc.msg, string(data))
		}
	}
}
