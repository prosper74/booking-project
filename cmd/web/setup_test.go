package main

// This file will run before any test file will run

import (
	"net/http"
	"os"
	"testing"
)

func TestMain(mainTest *testing.M) {
	// Before the test starts running, do something inside the `Run()` function, then exit
	os.Exit(mainTest.Run())
}

// `myHandler` holds objects that will satisfy http handler interface
type myHandler struct{}

func (handlerObject *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
