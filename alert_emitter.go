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

// Get results from the cache
func reportsHandler(w http.ResponseWriter, r *http.Request) {
	tmplName := "report.tmpl"
	// Temp
	tmplRoot := "/Users/yurii.rochniak/Repo/catchpoint_bridge/templates/"
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
