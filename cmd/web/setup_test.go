package main

import (
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type myHttpHandler struct{}

func (mh *myHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
