package main

import (
	"fmt"
	"github.com/yyamamot/groupssh"
)

func main() {

	group := groupssh.NewDotSshConfig([]groupssh.DotSshConfig{
		{Host: "host1"},
		{Host: "host2"},
		{Host: "host3"},
	})

	results := group.Get("/etc/hostname", "/tmp/{host}-{port}-hostname")
	for _, result := range results {
		if result.Error != nil {
			fmt.Print("Error on ", result.AltHost, ": ", result.Error)
		} else {
			fmt.Print("Output from ", result.AltHost, ": ", result.Stdout)
		}
	}
}

/*
$ go run ./examples/get.go
Output from host2: File /etc/hostname processed to /tmp/host2-232-hostname
Output from host1: File /etc/hostname processed to /tmp/host1-231-hostname
Output from host3: File /etc/hostname processed to /tmp/host3-233-hostname

$ ls /tmp/host*
/tmp/host1-231-hostname /tmp/host2-232-hostname /tmp/host3-233-hostname
*/

