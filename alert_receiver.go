package main

// This is a script that creates a web server to receive push calls from the
// Catchpoint Alerts API and sends it to nagios.
//
// Script parameters:
//   --verbose
//   --ip=X.X.X.X
//   --port=YYYY
//   --authorized-ips=X.X.X.X,y.y.y.y,Z.Z.Z.Z
//     Current IP of the Alerts API push servers: 64.79.149.6/32
//     Reference: https://support.catchpoint.com/hc/en-us/articles/202459889-Alerts-API
//
// Usage example:
//  - server (this application) side:
//    ./alert_receiver --verbose
//  - On the client side (to test):
//    CURLFORMAT='\ntime_namelookup:%{time_namelookup},\ntime_connect:%{time_connect},\ntime_appconnect:%{time_appconnect},\ntime_pretransfer:%{time_pretransfer},\ntime_redirect:%{time_redirect},\ntime_starttransfer:%{time_starttransfer},\ntime_total:%{time_total},\nnum_connects:%{num_connects},\nnum_redirects:%{num_redirects}\n'
//    curl  -X POST -d @/tmp/alert_api.xml http://127.0.0.1:8080/catchpoint --header "Content-Type:application/xml" -w "%${CURLFORMAT}"
//
// Recommendations:
//   Put this server behind a haproxy + an iptable that filter out all the
//   source IPs and rejects everything that is not on the correct endpoint
//   (here: /catchpoint) use the lb as a proxy and make the script listen on
//   127.0.0.1 only.
//

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/tubemogul/catchpoint_api_sdk_go/tree/master/alertAPI"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Those are the arguments you can pass to the script
var (
	// Flag to run or not the deamon in Verbose mode. Defaults to false.
	verbose = flag.Bool("verbose", false, "Set a verbose output")
	// Dumps the query to the log output
	// dumpRequests = flag.Bool("dump-requests", false, "Dumps the requests to the current log output")
	// Path to the configuration file. Defaults to "./receiver.cfg.json"
	configFile = flag.String("config", "./receiver.cfg.json", "Path to the config file")
)

var config = new(Configuration)

// checkIpFiltering sends an empty response if an IP filtering is defined and
// the IP is out of this filter.
func checkIpFiltering(clientIP *string) bool {
	if len(config.AuthIPs) > 0 {
		client_ip := strings.Split(*clientIP, ":")[0]
		for _, autorized_ip := range strings.Split(config.AuthIPs, ",") {
			if client_ip == autorized_ip {
				if *verbose {
					log.Printf("Accepted IP: %s", client_ip)
				}
				return true
			}
		}
		if *verbose {
			log.Printf("Refused IP: %s", *clientIP)
		}
		return false
	}
	return true
}

// verifyRequestContent checks if the content of the request is empty
func verifyRequestContent(w *http.ResponseWriter, req *http.Request) bool {

	if *verbose {
		log.Printf("Length of the query: %d", req.ContentLength)
	}
	// Check for a request body
	if req.ContentLength == 0 {
		http.Error(*w, http.StatusText(400), 400)
		return false
	}
	return true
}

// This function will take care of sending the nsca packets
// As of today, Go does not currently support the ciphersuite of nsca-ng so I'm
// using a call to the command line as an ugly workaround
func sendNscaMessage(state *uint8, service *string, message *string) error {
	if !config.NSCA.Enabled {
		return nil
	}

	cmd := exec.Command(config.NSCA.OsCommand, "-H", config.NSCA.Server, "-c", config.NSCA.ConfigFile)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%d\t%s\t%s", *state, *service, *message))
	err := cmd.Run()

	if *verbose {
		// In verbose mode, we print the command output before sending back the
		// error as both can be helpful for debug purposes.
		var out bytes.Buffer
		cmd.Stdout = &out
		log.Printf("NSCA command output: %s", out.String())
	}
	if err != nil {
		return err
	}

	return nil
}

// The handler that will redirect to the correct plugin
func genericHandler(w http.ResponseWriter, r *http.Request) {

	if *verbose {
		log.Printf("Receiving a new query from %s on %s", r.RemoteAddr, r.URL.Path)
	}

	// Doing nothing if the request is not from an authorized IP
	if !checkIpFiltering(&(r.RemoteAddr)) {
		return
	}

	// Doing nothing if the request is empty
	if !verifyRequestContent(&w, r) {
		return
	}

	var (
		rc  uint8
		svc *string
		msg *[]string
		err error
	)

	for _, endpoint := range config.Endpoints {
		if endpoint.URIPath == r.URL.Path {
			// Once you have the right endpoint, you check for the right plugin
			switch endpoint.PluginName {
			default:
				log.Printf("[ERROR] Unsupported plugin name for %s", endpoint.PluginName)
				return
			case "catchpoint":
				plugin := new(catchpoint.Alert)
				rc, svc, msg, err = plugin.RequestHandler(&(r.Body))
				if *verbose {
					log.Printf("RC = %d, Svc = %s,  Msg = %s, err = %s", rc, *svc, *msg, err)
				}
			}

			// Sending NSCA messages if enabled
			if config.NSCA.Enabled {
				// We send an nsca alert for each failures in the test to have a better
				// report of the frequency of the failures
				for _, failure := range *msg {
					err := sendNscaMessage(&rc, svc, &failure)
					if err != nil {
						log.Printf("[ERROR] %s", err.Error())
					}
				}
			}
			break // break when you find the matching endpoint
		}
	}
}

// Main function
func main() {
	flag.Parse()

	// load plugins

	// Loading the configuration
	if *verbose {
		log.Printf("Loading config")
	}
	err := config.loadConfig(*configFile)
	if err != nil {
		log.Fatal("Unable to laod configuration: %s", err)
	}

	// Multithreading the http server
	runtime.GOMAXPROCS(config.Procs)

	// Default route. We use it to handle every request. The filtering out is done
	// in the handler
	http.HandleFunc("/", genericHandler)

	if *verbose {
		log.Printf("Starting web server listening on %s:%d", config.IP, config.Port)
	}
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.IP, config.Port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	log.Fatal(s.ListenAndServe())
}
