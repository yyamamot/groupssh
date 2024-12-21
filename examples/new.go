package main

import (
	"fmt"
	"github.com/yyamamot/groupssh"
)

func main() {

	group := groupssh.New([]groupssh.SshConfig{
		{Host: "localhost", User: "root", Pass: "", Port: "231"},
		{Host: "localhost", User: "root", Pass: "", Port: "232"},
		{Host: "localhost", User: "root", Pass: "", Port: "233"},
	})

	results := group.Run("hostname")
	for _, result := range results {
		if result.Error != nil {
			fmt.Print("Error on ", result.Host, ": ", result.Error)
		} else {
			fmt.Print("Output from ", result.Host, ": ", result.Stdout)
		}
	}
}

/*
$ go run ./examples/new.go
Output from localhost: 77dd91fdf86f
Output from localhost: dad24c1d844c
Output from localhost: 427c640edd0d
*/
