package main

// Here we create the small function that help us to lighten the rest of the code. Aka: helpers!

import (
	"log"
	"net/http"
)

// handleErrorHttp handles the logging of error messages and returns and error
// 500 to the client
//
// Parameters:
// - e (*error): the error to print out.
// - w (*http.ResponseWriter): the http response object to write to
//
// Returns: nothing.
//
func handleErrorHttp(e *error, w *http.ResponseWriter) {
	if *e != nil {
		logError(e)
		http.Error(*w, http.StatusText(500), 500)
	}
}

// logError handles the logging of error messages and that's it.
//
// Parameters:
// - e (*error): the error to print out.
//
// Returns: nothing.
//
func logError(e *error) {
	if *e != nil {
		log.Printf("[ERROR] %s", (*e).Error())
	}
}

// logInfo handles the logging of informational messages and if the verbose mode
// is on.
//
// Parameters:
// - e (*error): the error to print out.
//
// Returns: nothing.
//
func logInfo(msg string) {
	if *verbose {
		log.Printf("[INFO] %s", msg)
	}
}
