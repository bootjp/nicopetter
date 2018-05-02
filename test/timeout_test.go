package main

import (
	"net/http"
	"testing"
	"time"
)

func TestTimeOutFail(t *testing.T) {
	c := http.Client{Timeout: time.Duration(500 * time.Millisecond)}
	_, err := c.Get("https://httpstat.us/200?sleep=5000")
	if err == nil {
		t.Fatal("timeout test fail.")
	}
}
