package main

import (
	"fmt"
	"github.com/yyamamot/groupssh"
)

func main() {
	/*
		$ cat ~/.ssh/config
		Host host1
		     HostName localhost
		     User root
		     Port 231

		Host host2
		     HostName localhost
		     User root
		     Port 232

		Host host3
		     HostName localhost
		     User root
		     Port 233
	*/
	group := groupssh.NewDotSshConfig([]groupssh.DotSshConfig{
		{Host: "host1"},
		{Host: "host2"},
		{Host: "host3"},
	})

	results := group.Run("hostname")
	for _, result := range results {
		if result.Error != nil {
			fmt.Print("Error on ", result.AltHost, ": ", result.Error)
		} else {
			fmt.Print("Output from ", result.AltHost, ": ", result.Stdout)
		}
	}
}

/*
$ go run ./examples/newdotsshcofnig.go
Output from host3: 427c640edd0d
Output from host2: dad24c1d844c
Output from host1: 77dd91fdf86f
*/
