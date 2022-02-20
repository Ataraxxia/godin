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
	if string(data) != "Godin" {
		t.Errorf("expected Godin got %v", string(data))
	}
}

func TestUploadReport(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	const testdatapath string = "testdata/godin/"
	testtable := []struct {
		file     string
		expected bool
	}{
		{
			file:     "apt_report_ok_1.json",
			expected: true,
		},
	}

	db = MockDB{
		initializeDatabaseErr: nil,
		saveReportErr:         nil,
	}

	for _, tc := range testtable {
		f, err := os.Open(testdatapath + tc.file)
		if err != nil {
			t.Fatalf("Could not open test file %s", tc.file)
		}
		req := httptest.NewRequest(http.MethodPost, "/reports/upload", f)
		w := httptest.NewRecorder()

		uploadReport(w, req)

		res := w.Result()
		defer res.Body.Close()

		data, err := ioutil.ReadAll(res.Body)

		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}

		if string(data) != "Godin says OK\n" {
			t.Errorf("expected 'Godin says ok', got %v", err)
		}
	}
}
