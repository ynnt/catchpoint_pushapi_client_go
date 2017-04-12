package main

// This is a script that create a web server, which responds
// with results of Catchpoint checks, which been stored in the cache
// Output format: JSON
// JSON is compatible with tm-health-check format

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"text/template"
)

// Converts string to JSON format
func ToJSONString(v interface{}) string {
	bytesOutput, _ := json.Marshal(v)
	return string(bytesOutput)
}

// verifyHealthRequest checks if GET method is used for check request.
// If not, returns an HTTP error 400.
func verifyHealthRequest(w *http.ResponseWriter, req *http.Request) bool {

	logInfo(fmt.Sprintf("Health request. Method: %q. Length: %d", req.Method, req.ContentLength))

	if req.Method != "GET" {
		http.Error(*w, http.StatusText(400), 400)
		return false
	}
	return true
}

// Get results from the cache
func reportsHandler(w http.ResponseWriter, r *http.Request) {

	// Doing nothing if the request is not from an authorized IP
	if !checkIpFiltering(&(r.RemoteAddr)) {
		return
	}

	// Doing nothing if not a GET request
	if !verifyHealthRequest(&w, r) {
		return
	}

	tmplName := config.Emitter.Template
	tmplRoot := config.Emitter.TemplateDir
	tmplPath := filepath.Join(tmplRoot, tmplName)
	w.Header().Set("Content-Type", "application/json")
	fMaps := template.FuncMap{"tojson": ToJSONString}
	t := template.Must(template.New(tmplName).Funcs(fMaps).ParseFiles(tmplPath))
	io.WriteString(w, "[")
	i := 1
	for host, svcs := range cache {
		j := 1
		for svc, chk := range svcs {
			c := map[string]map[string]interface{}{
				"check": map[string]interface{}{"host": host, "name": svc, "status": chk.state, "message": chk.output, "timestamp": fmt.Sprint(chk.timestamp), "statusFirstSeen": fmt.Sprint(chk.statusFirstSeen)},
			}
			t.Execute(w, c)
			// This part just takes care of adding a coma or not between the elements
			// to have a correcly-formated json
			if !(i == len(cache) && j == len(svcs)) {
				io.WriteString(w, ",")
			}
			j++
		}
		i++
	}
	io.WriteString(w, "]\n")
	logInfo(fmt.Sprintf("Items were read from the cache: %d", len(cache)))
}
