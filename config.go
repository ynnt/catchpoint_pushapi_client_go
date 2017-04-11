package main

// Configuration management for the http server part

import (
	"encoding/json"
	"os"
)

// Main configuration struct
type Configuration struct {
	// IP on which the webserver should listen on. By default it's 127.0.0.1.
	// Name in the config file: listener_ip
	IP string `json:"listener_ip"`

	// Port on which the webserver should listen on. By default it's 8080.
	// Name in the config file: listener_port
	Port uint16 `json:"listener_port"`

	// Comma-separated list of IP to only listen from. Queries from other
	// IPs will be discarded. Empty to serve all the clients without restrictions
	// which is the default value.
	// For the list of the current IP Catchpoint is pushing from, see:
	// https://support.catchpoint.com/hc/en-us/articles/202459889-Alerts-API
	// Name in the config file: authorized_ips
	AuthIPs string `json:"authorized_ips"`

	// The number of concurrent CPUs the HTTP server is allowed to use.
	// This sets GOMAXPROCS at the run time.
	// If the GOMAXPROCS environment variable is set and to a bigger value than
	// this number, the value of the environment variable will be taken.
	// For more details, see https://golang.org/src/runtime/debug.go?s=995:1021#L16
	// By default this configuration value is set to 1.
	// Name in the config file: max_procs
	Procs int `json:"max_procs"`

	// Path to the log file you want to have the output to. Keep it empty if you
	// want to log to the console.
	// Defaults to empty so console logging.
	LogFile string `json:"log_file"`

	// Endpoints list specifying which plugin should handle which endpoint
	Endpoints []Endpoint `json:"endpoints"`

        // Emitter. Listener to give results back
        Emitter []Emitter `json:"emitter"`

	// Configuration of the nsca plugin
	NSCA Nsca `json:"nsca"`
}

// Endpoint from which results being gathered
type Emitter struct {
    URIPath string `json:"uri_path"`
}

// The endpoints define which plugin is used for each supported endpoint
type Endpoint struct {
	// The definition of the Path of the endpoint (for example "/catchpoint/alerts")
	URIPath string `json:"uri_path"`

	// The name of the plugin that is supposed to handle this endpoint.
	// Currently supported values:
	//   - catchpoint_alerts
	PluginName string `json:"plugin_name"`
}

// Configuration of NSCA
type Nsca struct {
	// Wether or not we want to send alerts with this method. If empty the default
	// value will be false (as it is the default value of a boolean in Go)
	Enabled bool `json:"enabled"`

	// The name of the NSCA server to send the data to. No default value.
	Server string `json:"server"`

	// Full path of the send_nsca command on the system
	// Defaults to "/usr/sbin/send_nsca"
	OsCommand string `json:"os_command_path"`

	// Configuration file path for the send_nsca command
	// Defaults to "/etc/send_nsca.cfg"
	ConfigFile string `json:"config_file"`

	// The name of the host you want to use when sending the nsca messages
	ClientHost string `json:"client_host"`
}

// Configuration for Sensu
type Sensu struct {
	// Status of the check
	Status uint8 `json:"status"`

	// Name of the service
	Name string `json:"name"`

	// Message
	Output string
}

// This function loads the configuration file given in parameter and returns a
// pointer to a Configuration object
func (cfg *Configuration) loadConfig(confFilePath string) error {
	file, oserr := os.Open(confFilePath)
	if oserr != nil {
		return oserr
	}

	decoder := json.NewDecoder(file)
	err := decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	// Loading the default configurations
	if cfg.IP == "" {
		cfg.IP = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.Procs == 0 {
		cfg.Procs = 1
	}
	if len(cfg.NSCA.ClientHost) == 0 {
		cfg.NSCA.ClientHost, err = os.Hostname()
		return err
	}
	if len(cfg.NSCA.OsCommand) == 0 {
		cfg.NSCA.OsCommand = "/usr/sbin/send_nsca"
	}
	if len(cfg.NSCA.ConfigFile) == 0 {
		cfg.NSCA.ConfigFile = "/etc/send_nsca.cfg"
	}

	return nil
}
