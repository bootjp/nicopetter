package main

import (
	"errors"
	"net/http"
	"testing"
)

type httpTestCase struct {
	url       string
	wantError bool
}

var httpTestData = []httpTestCase{
	{"https://httpstat.us/500", true},
	{"https://httpstat.us/502", true},
	{"https://httpstat.us/503", true},
	{"https://httpstat.us/403", true},
	{"https://httpstat.us/400", true},
	{"https://httpstat.us/401", true},
	{"https://httpstat.us/302", false},
}

func TestHTTPStatus(t *testing.T) {

	for _, v := range httpTestData {
		if v.wantError && errorRes(v.url) == nil {
			t.FailNow()
		}
	}

}

func errorRes(url string) error {

	c := http.Client{}

	r, err := c.Get(url)
	defer r.Body.Close()
	if err != nil {
		return err
	}
	switch r.Status[:1] {
	case "4", "5":
		return errors.New("got 40x or 50x status code.")

	case "3":
		return errors.New("warn  got 30x statsu code.")

	}
	return nil

}
