package pcc

import (
	log "github.com/platinasystems/go-common/logs"

	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strings"
)

type SSHHandler struct {
}

type SshConfiguration struct {
	User            string
	Key             string
	Port            int
	SshClientConfig *ssh.ClientConfig
}

var sshConfig *SshConfiguration

func InitSSH(cfg *SshConfiguration) {
	if cfg == nil {
		return
	}
	sshConfig = cfg

	if len(strings.TrimSpace(sshConfig.User)) > 0 {
		log.AuctaLogger.Infof("init ssh handler", sshConfig.User, sshConfig.Key)

		if sshConfig.Port <= 0 {
			sshConfig.Port = 22
		}

		if buffer, err := ioutil.ReadFile(sshConfig.Key); err == nil {
			if signer, err := ssh.ParsePrivateKey(buffer); err == nil {
				sshConfig.SshClientConfig = &ssh.ClientConfig{
					User:            sshConfig.User,
					Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				}

			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}

func (sh *SSHHandler) Run(host string, command string) (stdout string, stderr string, err error) {
	var (
		client  *ssh.Client
		session *ssh.Session
	)

	log.AuctaLogger.Infof(fmt.Sprintf("Running command [%s] on host %s", command, host))
	if client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, sshConfig.Port), sshConfig.SshClientConfig); err == nil {
		defer client.Close()

		if session, err = client.NewSession(); err == nil {
			defer session.Close()
			var (
				stdoutBuf bytes.Buffer
				stderrBuf bytes.Buffer
			)

			session.Stdout = &stdoutBuf
			session.Stderr = &stderrBuf
			err = session.Run(command)

			stdout = stdoutBuf.String()
			stderr = stderrBuf.String()
		}
	}
	return
}

type Cmd struct {
	Input      string
	Stdout     string
	Stderr     string
	SessionErr error
}

func (sshC *SshConfiguration) Init() (err error) {
	var (
		buffer []byte
		signer ssh.Signer
	)
	if len(strings.TrimSpace(sshC.User)) == 0 {
		err = fmt.Errorf("No ssh user found in config")
		return
	}
	if sshC.Port <= 0 {
		sshC.Port = 22
	}

	if buffer, err = ioutil.ReadFile(sshC.Key); err == nil {
		if signer, err = ssh.ParsePrivateKey(buffer); err == nil {
			sshC.SshClientConfig = &ssh.ClientConfig{
				User:            sshC.User,
				Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}
		}
	}

	return
}

func (sshC *SshConfiguration) Run(host string, cmds ...*Cmd) (err error) {
	var (
		client *ssh.Client
	)

	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	if sshC.SshClientConfig == nil {
		err = fmt.Errorf("sshConfig not initialized")
		return
	}
	if client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, sshC.Port), sshC.SshClientConfig); err != nil {
		return
	}

	for _, cmd := range cmds {
		var (
			stdoutBuf bytes.Buffer
			stderrBuf bytes.Buffer
		)
		session, e := client.NewSession()
		if session != nil {
			defer session.Close()
		}
		if e != nil {
			cmd.SessionErr = e
			continue
		}

		session.Stdout = &stdoutBuf
		session.Stderr = &stderrBuf
		err = session.Run(cmd.Input)

		cmd.Stdout = stdoutBuf.String()
		cmd.Stderr = stderrBuf.String()
	}

	return
}
