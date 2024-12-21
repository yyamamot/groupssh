package groupssh

import (
	"bytes"
	"fmt"
	"github.com/kevinburke/ssh_config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DotSshConfig .ssh/config configuration
type DotSshConfig struct {
	Host    string        // Hostname or IP address
	Pass    string        // Password (default ""), if IdentityFile is set, does not use password.
	Timeout time.Duration // Timeout (default 0: no timeout)
}

// SshConfig user specific configuration
type SshConfig struct {
	Host    string        // Hostname or IP address
	AltHost string        // Alternative Hostname (.ssh/config HostName)
	User    string        // Username (default "root")
	Pass    string        // Password (default ""), Pass or KeyPath must be set
	KeyPath string        // Path to private key (default ""), if set, does not use password.
	Port    string        // Port (default 22)
	Timeout time.Duration // Timeout (default 0: no timeout)
}

func (c *SshConfig) sshAddr() string {
	return c.Host + ":" + c.Port
}

func (c *SshConfig) clientConfig() *ssh.ClientConfig {
	// password auth
	if c.KeyPath == "" {
		return &ssh.ClientConfig{
			User: c.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.Pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         c.Timeout,
		}
	}

	// key auth
	buf, err := os.ReadFile(c.KeyPath)
	if err != nil {
		panic(err)
	}
	key, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		panic(err)
	}
	return &ssh.ClientConfig{
		User: c.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		Timeout: c.Timeout,
	}
}

func NewDotSshConfig(sshConfigs []DotSshConfig) *GroupSsh {
	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var configs []SshConfig
	for _, sshConfig := range sshConfigs {
		// Seek to the beginning of the file to ensure ssh_config.Decode works correctly
		_, err = f.Seek(0, 0)
		if err != nil {
			panic(err)
		}

		cfg, err := ssh_config.Decode(f)
		if err != nil {
			panic(err)
		}

		user, _ := cfg.Get(sshConfig.Host, "User")
		if user == "" {
			user = "root"
		}
		port, _ := cfg.Get(sshConfig.Host, "Port")
		if port == "" {
			port = "22"
		}
		hostName, _ := cfg.Get(sshConfig.Host, "HostName")
		if hostName == "" {
			hostName = sshConfig.Host
		}
		keyPath, _ := cfg.Get(sshConfig.Host, "IdentityFile")

		configs = append(configs, SshConfig{
			Host:    hostName,
			AltHost: sshConfig.Host,
			User:    user,
			Pass:    sshConfig.Pass,
			KeyPath: keyPath,
			Port:    port,
			Timeout: sshConfig.Timeout,
		})
	}

	return &GroupSsh{configs: configs}
}

func New(configs []SshConfig) *GroupSsh {
	return &GroupSsh{configs: configs}
}

type GroupSsh struct {
	configs []SshConfig
}

type Result struct {
	Host    string
	AltHost string
	Stdout  string
	Stderr  string
	Error   error
}

func (p *GroupSsh) connectAndExecute(config SshConfig, remote, local string, action func(*sftp.Client, string, string) error) Result {
	client, err := ssh.Dial("tcp", config.sshAddr(), config.clientConfig())
	if err != nil {
		return Result{Host: config.Host, AltHost: config.AltHost, Error: err}
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return Result{Host: config.Host, AltHost: config.AltHost, Error: err}
	}
	defer sftpClient.Close()

	err = action(sftpClient, remote, local)
	if err != nil {
		return Result{Host: config.Host, AltHost: config.AltHost, Error: err}
	}

	return Result{Host: config.Host, AltHost: config.AltHost, Stdout: fmt.Sprintf("File %s processed to %s\n", remote, local)}
}

func (p *GroupSsh) getFile(config SshConfig, remote, local string) Result {
	return p.connectAndExecute(config, remote, local, func(sftpClient *sftp.Client, remote, local string) error {
		remoteFile, err := sftpClient.Open(remote)
		if err != nil {
			return err
		}
		defer remoteFile.Close()

		localFile, err := os.Create(local)
		if err != nil {
			return err
		}
		defer localFile.Close()

		_, err = io.Copy(localFile, remoteFile)
		return err
	})
}

func (p *GroupSsh) putFile(config SshConfig, remote, local string) Result {
	return p.connectAndExecute(config, remote, local, func(sftpClient *sftp.Client, remote, local string) error {
		localFile, err := os.Open(local)
		if err != nil {
			return err
		}
		defer localFile.Close()

		remoteFile, err := sftpClient.Create(remote)
		if err != nil {
			return err
		}
		defer remoteFile.Close()

		_, err = io.Copy(remoteFile, localFile)
		return err
	})
}

func (p *GroupSsh) Get(remote, local string) []Result {
	var wg sync.WaitGroup
	resultChan := make(chan Result, len(p.configs))

	for _, config := range p.configs {
		wg.Add(1)
		go func(config SshConfig) {
			defer wg.Done()

			// replace {host} and {port} in local
			host := config.Host
			if config.AltHost != "" {
				host = config.AltHost
			}
			hostReplacedLocal := strings.Replace(local, "{host}", host, -1)
			portReplacedLocal := strings.Replace(hostReplacedLocal, "{port}", config.Port, -1)

			result := p.getFile(config, remote, portReplacedLocal)
			resultChan <- result
		}(config)
	}

	wg.Wait()
	close(resultChan)

	var results []Result
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func (p *GroupSsh) Put(local, remote string) []Result {
	var wg sync.WaitGroup
	resultChan := make(chan Result, len(p.configs))

	for _, config := range p.configs {
		wg.Add(1)
		go func(config SshConfig) {
			defer wg.Done()

			result := p.putFile(config, remote, local)
			resultChan <- result
		}(config)
	}

	wg.Wait()
	close(resultChan)

	var results []Result
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func (p *GroupSsh) executeCommand(config SshConfig, command string) Result {
	client, err := ssh.Dial("tcp", config.sshAddr(), config.clientConfig())
	if err != nil {
		if os.IsTimeout(err) {
			return Result{Host: config.Host, AltHost: config.AltHost, Error: fmt.Errorf("connection timed out: %w", err)}
		}
		return Result{Host: config.Host, AltHost: config.AltHost, Error: err}
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return Result{Host: config.Host, AltHost: config.AltHost, Error: err}
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(command)
	if err != nil {
		return Result{Host: config.Host, AltHost: config.AltHost, Error: fmt.Errorf("failed to create session: %w", err)}
	}
	defer session.Close()

	return Result{
		Host:    config.Host,
		AltHost: config.AltHost,
		Stdout:  stdoutBuf.String(),
		Stderr:  stderrBuf.String(),
		Error:   err,
	}
}

func (p *GroupSsh) Run(command string) []Result {
	var wg sync.WaitGroup
	resultChan := make(chan Result, len(p.configs))

	for _, config := range p.configs {
		wg.Add(1)
		go func(config SshConfig) {
			defer wg.Done()
			result := p.executeCommand(config, command)
			resultChan <- result
		}(config)
	}

	wg.Wait()
	close(resultChan)

	var results []Result
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}
