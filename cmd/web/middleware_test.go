package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var myHandler *myHttpHandler

	noSurfHandler := NoSurf(myHandler)

	switch v := noSurfHandler.(type) {
	case http.Handler:
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var myHandler *myHttpHandler

	loadAndSaveHandler := app.Session.LoadAndSave(myHandler)

	switch v := loadAndSaveHandler.(type) {
	case http.Handler:
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %T", v))
	}
}
