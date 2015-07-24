# A client for the Catchpoint Push Alert API written in Go

## Installing and running

### Requirements

The API requires Go 1.2 or higher to be installed on the server.

### Installation script example

In this example, we'll install the application in /usr/local/go/bin. If you want
to install it somewhere else, feel free to change the GOPATH.
It also uses the hostname as client host (nagios "host" where the checks will be
grouped), but you might want to change it to something more relevant.

This example supposes that you have a version of go 1.2+ installed on the
machine before you execute this. (Git is also required here as we use the `go
get` command which uses it)

```
export GOPATH=/usr/local/go
export LOGROOT=/var/logs
mkdir -p $GOPATH
go get github.com/tubemogul/catchpoint_pushapi_client_go
go install github.com/tubemogul/catchpoint_pushapi_client_go

# Config example
cat << __EOF__ > /etc/catchpoint_pushapi_client.cfg
{
  "listener_ip": "0.0.0.0",
  "listener_port": 80,
  "authorized_ips": "64.79.149.6",
  "max_procs": 4,
  "log_file": "${LOGROOT}/catchpoint.log",
  "endpoints":[
    { "uri_path": "/catchpoint/alerts",
      "plugin_name": "catchpoint_alerts"}
  ],
  "nsca": {
    "enabled": true,
    "server": "nsca_server.example.com",
    "os_command_path": "/usr/sbin/send_nsca",
    "config_file": "/etc/send_nsca.cfg",
    "client_host": "$(hostname)"
  }
}
__EOF__

# In this example we'll dump all the incoming body requests
mkdir -p ${LOGROOT}/catchpoint/

# Yes, an init script would be way better, it's in my TODO! :)
nohup ${GOPATH}/bin/catchpoint_pushapi_client_go \
  --verbose \
  --config=/etc/receiver.cfg.json \
  --dump-requests-dir=${LOGROOT}/catchpoint/ &
```

## Contributing

See CONTRIBUTING.md file.

## License

This script is distributed under the Apache License, Version 2.0, see LICENSE file
for more informations.
