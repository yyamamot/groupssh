# groupssh 

- groupssh is a golang library for defining groups and making bulk ssh connections to those groups.
- groupssh was inspired by the Group feature of [fabric](https://github.com/fabric/fabric).
- Since this is a single-file library, please copy and customize the necessary parts according to your environment.

## API

- New(configs []SshConfig)
- NewDotSshConfig(configs []DotSshConfig)
- Run(command string)
- Put(local, remote string)
- Get(remote, local string)
  - The `local` variable supports the following placeholders:
     - {host}: Replaced with the hostname of the destination
     - {port}: Replaced with the port number of the destination
## Example

### New(): create a new groupssh instance

```go
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
```

### NewDotSshConfig(): create a new groupssh instance with the ssh config file

```go
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
```

### Put(): copy a file from local to all hosts

```go
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
```

### Get(): copy a file from all hosts to local

```go
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
```
