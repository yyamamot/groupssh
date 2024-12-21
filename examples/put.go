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

	results := group.Put("hello", "/tmp/hello")
	for _, result := range results {
		if result.Error != nil {
			fmt.Print("Error on ", result.AltHost, ": ", result.Error)
		} else {
			fmt.Print("Output from ", result.AltHost, ": ", result.Stdout)
		}
	}
}

/*
$ go run ./examples/put.go
Output from host3: File /tmp/hello processed to hello
Output from host2: File /tmp/hello processed to hello
Output from host1: File /tmp/hello processed to hello

# host1
root@77dd91fdf86f:/app# cat /tmp/hello
hello
*/
