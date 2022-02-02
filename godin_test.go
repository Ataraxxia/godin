package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"testing"
)

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
