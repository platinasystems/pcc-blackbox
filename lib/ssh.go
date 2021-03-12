package pcc

import (
	"bytes"
	"fmt"
	log "github.com/platinasystems/go-common/logs"
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
