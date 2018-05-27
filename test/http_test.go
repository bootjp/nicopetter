package main

import (
	"errors"
	"log"
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
	{"https://httpstat.us/204", false},
}

func TestHTTPStatus(t *testing.T) {

	for _, v := range httpTestData {
		res := errorRes(v.url)
		switch v.wantError {
		case true:
			if res == nil {
				t.FailNow()
			}
		case false:
			if res != nil {
				t.FailNow()
			}
		}
	}

}

func errorRes(url string) error {

	c := http.Client{}

	r, err := c.Get(url)
	if err != nil {
		return err
	}

	defer func() {
		err = r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	switch r.Status[:1] {
	case "4", "5":
		return errors.New("got 40x or 50x status code")
	}

	return err
}
