package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(testPointer *testing.T) {
	var handlerObject myHandler
	testHandler := NoSurf(&handlerObject)

	// return the type of the handler 
	switch handlerType := testHandler.(type) {
	case http.Handler:
		// do nothing
	default:
		testPointer.Error(fmt.Sprintf("type is not http.Handler but is %T", handlerType))
	}
}

func TestSessionLoad(testPointer *testing.T) {
	var handlerObject myHandler
	testHandler := SessionLoad(&handlerObject)

	switch handlerType := testHandler.(type) {
	case http.Handler:
		// do nothing
	default:
		testPointer.Error(fmt.Sprintf("type is not http.Handler but is %T", handlerType))
	}
}
