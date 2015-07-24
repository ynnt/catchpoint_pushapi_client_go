package main

// Here we will set only the elements related to handling the NSCA calls

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// This function will take care of sending the nsca packets
// As of today, Go does not currently support the ciphersuite of nsca-ng so I'm
// using a call to the command line as an ugly workaround.
//
// Parameters:
// - state (*uint8): the nsca status to send (0, 1, 2, 3)
// - service (*string): the name of the nagios service that will handle the
//   request. The current plugins contruct the service using like:
//     Product name-Test name
// - message (*string): the message describing the alert to send
//
// Returns:
// - error: an error if one has been encountered
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
